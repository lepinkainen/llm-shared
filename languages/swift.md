# Swift Guidelines

For SwiftUI framework behavior — hosting/constraint crashes, layout,
state/observation, lists, typography, windows, testing strategy — see
[`swiftui.md`](swiftui.md). This file covers Swift-the-language, AppKit
menu-bar apps, packaging, and formatting with SwiftFormat.

## macOS Menu Bar (Tray) Applications

### Entry Point — Use MenuBarExtra

**Default to the `MenuBarExtra` scene.** It is the supported SwiftUI way to put
an item in the menu bar on macOS 13+, and it owns the status item and the
anchoring of its content for you. Use `.menuBarExtraStyle(.window)` when the
content is a rich SwiftUI view, or `.menu` for a plain dropdown menu.

```swift
@main
struct MyApp: App {
    @NSApplicationDelegateAdaptor(AppDelegate.self) var appDelegate

    var body: some Scene {
        MenuBarExtra {
            PopoverContentView(viewModel: appDelegate.viewModel)
        } label: {
            MenuBarLabel(viewModel: appDelegate.viewModel)
        }
        .menuBarExtraStyle(.window)

        Settings {
            SettingsView(viewModel: appDelegate.viewModel)
        }
    }
}
```

Keep `@NSApplicationDelegateAdaptor` alongside it for the things that still need
an `AppDelegate` — owning a long-lived view model, opening sockets, registering
for notifications. It is the *manual status item* that `MenuBarExtra` replaces,
not the delegate.

Add a `Settings` scene for preferences; SwiftUI wires it to the standard
Settings menu item and keyboard shortcut. Set `LSUIElement` in `Info.plist` to
keep the app out of the Dock, which is the declarative equivalent of calling
`NSApp.setActivationPolicy(.accessory)`.

Never use the manual `NSApplication.shared.run()` pattern:

```swift
// WRONG — causes popover misplacement
// let app = NSApplication.shared
// app.setActivationPolicy(.accessory)
// let delegate = AppDelegate()
// app.delegate = delegate
// app.run()
```

### AppDelegate Pattern (manual status item)

**Prefer `MenuBarExtra` above.** Build the status item and popover by hand only
when you need something the scene does not expose — custom anchoring or
positioning, a target older than macOS 13, or click behavior that has to differ
per mouse button. The detached-popover bug this section used to warn about is a
consequence of driving the anchoring yourself; `MenuBarExtra` cannot hit it.

If you do go manual, the key points are:

- Create the `NSStatusItem` first, then set activation policy to `.accessory`
- Use `NSPopover` with `.transient` behavior for auto-dismiss on click-away
- Wrap SwiftUI views in `NSHostingController` for the popover content
- Call `NSApp.activate(ignoringOtherApps: true)` before showing the popover

```swift
class AppDelegate: NSObject, NSApplicationDelegate {
    var statusItem: NSStatusItem!
    var popover: NSPopover!

    func applicationDidFinishLaunching(_ notification: Notification) {
        // Create status bar item first
        statusItem = NSStatusBar.system.statusItem(withLength: NSStatusItem.squareLength)

        if let button = statusItem.button {
            button.image = NSImage(systemSymbolName: "icon.name", accessibilityDescription: "My App")
            button.action = #selector(togglePopover)
            button.target = self
        }

        // Hide from Dock — after status item creation
        NSApp.setActivationPolicy(.accessory)

        // Create popover with SwiftUI content
        let popover = NSPopover()
        popover.contentSize = NSSize(width: 320, height: 480)
        popover.behavior = .transient
        popover.contentViewController = NSHostingController(rootView: ContentView())
        self.popover = popover
    }

    @objc func togglePopover() {
        guard let button = statusItem.button else { return }
        if popover.isShown {
            popover.performClose(nil)
        } else {
            NSApp.activate(ignoringOtherApps: true)
            popover.show(relativeTo: button.bounds, of: button, preferredEdge: .minY)
            popover.contentViewController?.view.window?.makeKey()
        }
    }
}
```

### SwiftUI + ObservableObject for Reactive State

When the data source is push-based (e.g., socket, timer), use an `ObservableObject` view model as the bridge:

```swift
class MyViewModel: ObservableObject {
    @Published var status: String = ""
    @Published var items: [Item] = []

    // Action callbacks — set by AppDelegate, invoked by SwiftUI views
    var onAction: (() -> Void)?

    func update(state: SomeState) {
        status = state.status
        items = state.items
    }
}
```

For toggles that send commands to an external process (where the real state comes back asynchronously), use `Binding(get:set:)` so the UI doesn't write local state directly:

```swift
Toggle("My Toggle", isOn: Binding(
    get: { viewModel.someFlag },
    set: { _ in viewModel.onToggleFlag?() }  // sends command; state comes back via next update
))
```

### App Icons for SPM-Based Apps

macOS uses the app bundle's icon (from `Contents/Resources/`) for notifications, Activity Monitor, and other system UI. `NSApp.applicationIconImage` is a runtime-only property — it does **not** affect notifications.

**Steps:**

1. **Generate `.icns` from a source PNG** (ideally 1024x1024):

```bash
mkdir -p AppIcon.iconset
for size in 16 32 64 128 256 512 1024; do
    sips -z $size $size source.png --out AppIcon.iconset/icon_${size}x${size}.png
done
# Create @2x variants
cp AppIcon.iconset/icon_32x32.png   AppIcon.iconset/icon_16x16@2x.png
cp AppIcon.iconset/icon_64x64.png   AppIcon.iconset/icon_32x32@2x.png
cp AppIcon.iconset/icon_256x256.png AppIcon.iconset/icon_128x128@2x.png
cp AppIcon.iconset/icon_512x512.png AppIcon.iconset/icon_256x256@2x.png
cp AppIcon.iconset/icon_1024x1024.png AppIcon.iconset/icon_512x512@2x.png
rm AppIcon.iconset/icon_64x64.png AppIcon.iconset/icon_1024x1024.png
iconutil -c icns AppIcon.iconset -o AppIcon.icns
```

2. **Add `CFBundleIconFile` to `Info.plist`** (without the `.icns` extension):

```xml
<key>CFBundleIconFile</key>
<string>AppIcon</string>
```

3. **Copy the `.icns` into the `.app` bundle** during packaging:

```yaml
# Taskfile / build script
- cp path/to/AppIcon.icns build/MyApp.app/Contents/Resources/
```

4. **Re-sign the app** after modifying the bundle:

```bash
codesign --sign - --force build/MyApp.app
```

5. **If — and only if — a launchd plist starts the app, launch it via `open -W`.**
   For autostart, prefer `SMAppService` (see *Launch at login* below), which has
   no plist and no bundle-association problem. When a plist is genuinely
   required and the binary is launched directly (`Contents/MacOS/MyApp`), macOS
   doesn't associate the process with the `.app` bundle, so the bundle icon
   won't appear in notifications:

```xml
<!-- CORRECT — LaunchServices associates process with .app bundle -->
<key>ProgramArguments</key>
<array>
    <string>/usr/bin/open</string>
    <string>-W</string>
    <string>/Applications/MyApp.app</string>
</array>

<!-- WRONG — direct binary launch, no bundle association -->
<key>ProgramArguments</key>
<array>
    <string>/Applications/MyApp.app/Contents/MacOS/MyApp</string>
</array>
```

The `-W` flag makes `open` wait for the app to quit, so `KeepAlive` in launchd still works correctly.

**Common mistakes:**
- Launching the binary directly from launchd instead of using `open -W` — the process won't be associated with the `.app` bundle
- Using `NSApp.applicationIconImage` — this only affects the Dock icon at runtime, not notifications
- Using `UNNotificationAttachment` — this adds media content to the notification body, it does not set the app icon badge
- Forgetting to re-sign after adding the icon to the bundle

### Launch at Login — Use SMAppService

For "start when I log in", register the app as a login item from inside the app
with `SMAppService.mainApp` (macOS 13+). It replaces shipping a LaunchAgent
plist: nothing to install into `~/Library/LaunchAgents`, nothing to `launchctl
load`, and the user can revoke it from System Settings › General › Login Items.

```swift
// Register / unregister
try SMAppService.mainApp.register()
try SMAppService.mainApp.unregister()

// Authoritative state — re-read it after every toggle rather than
// tracking a local Bool, since the user can change it in System Settings.
let enabled = SMAppService.mainApp.status == .enabled
```

Put it behind a small protocol so the surrounding view model can be unit-tested
without touching the real service.

A launchd plist is still the right tool for a **background daemon that is not an
app bundle** — a plain executable with no UI. Use `SMAppService` for the app,
launchd for the daemon, and don't mix the two for the same process.

### Platform & Package Manager

- Use Swift Package Manager (SPM) with `Package.swift`
- Target `macOS(.v13)` or later for full SwiftUI support
- No extra dependencies needed for menu bar apps — SwiftUI and AppKit are system frameworks
- For bundled resources (icons), use `.process("Resources")` in the SPM target and access via `Bundle.module`
- `swift-tools-version: 6.0` or later already defaults to Swift 6 language mode,
  so strict concurrency is on without any extra setting. Xcode projects are not
  covered by that default and still need `SWIFT_STRICT_CONCURRENCY = complete`
  in their build settings.

## Formatting Swift with SwiftFormat

Format Swift with [SwiftFormat](https://github.com/nicklockwood/SwiftFormat)
using the Airbnb style guide config. There is no SwiftLint in this setup —
SwiftFormat is the whole story.

Copy `templates/airbnb.swiftformat` into the project next to the Swift sources
and wire up two tasks:

```yaml
vars:
  SWIFTFORMAT_ARGS: --config path/to/airbnb.swiftformat path/to/Sources path/to/Tests

tasks:
  lint-apple:
    desc: Check Swift formatting (Airbnb style guide via SwiftFormat)
    platforms: [darwin]
    cmds:
      - swiftformat --lint {{.SWIFTFORMAT_ARGS}}

  format-apple:
    desc: Apply Airbnb SwiftFormat style to Swift sources
    platforms: [darwin]
    cmds:
      - swiftformat {{.SWIFTFORMAT_ARGS}}
```

Call the lint task from `lint` behind a Darwin guard so Linux builds skip it:

```yaml
  lint:
    cmds:
      - task: lint-go
      - |
        if [ "$(uname -s)" = "Darwin" ]; then
          task lint-apple
        fi
```

### Pin the version in three places

SwiftFormat's output changes between releases, so an unpinned formatter means
one machine reformats what another just formatted. The version appears in three
places and they must move together:

1. `--minversion` in the `.swiftformat` config
2. `mise.toml`, so local shells get that exact binary
3. the download URL in the CI workflow

```toml
# mise.toml — keep in step with --minversion and the CI download
[tools]
swiftformat = "0.63.0"
```

### CI: install the pinned release, not brew

macOS runner images ship their own SwiftFormat in `/opt/homebrew/bin`, and brew
will not upgrade an already-installed formula — so `brew install swiftformat`
silently leaves you on whatever the image happened to have. Download the pinned
release and put it *ahead* of the preinstalled one on `PATH`:

```yaml
  apple:
    runs-on: macos-26
    steps:
      - uses: actions/checkout@v7

      - name: Install Task
        uses: go-task/setup-task@v2
        with:
          version: 3.x

      - name: Install SwiftFormat
        run: |
          curl -sSL -o "$RUNNER_TEMP/swiftformat.zip" https://github.com/nicklockwood/SwiftFormat/releases/download/0.63.0/swiftformat.zip
          mkdir -p "$RUNNER_TEMP/swiftformat"
          unzip -q -o "$RUNNER_TEMP/swiftformat.zip" -d "$RUNNER_TEMP/swiftformat"
          # Must precede the runner's preinstalled /opt/homebrew/bin/swiftformat
          echo "$RUNNER_TEMP/swiftformat" >> "$GITHUB_PATH"

      - name: Toolchain versions
        run: |
          swiftformat --version
          swift --version

      - name: Lint Swift
        run: task lint-apple

      - name: Test
        run: task test-apple

      - name: Build
        run: task build-apple
```

Swift needs its own job because it needs a macOS runner — the Linux jobs in a
mixed-language repo never build it, so without this job nothing enforces the
formatting at all.

### Tests: prefer swift-testing

Write unit tests with swift-testing (`import Testing`, `@Test`, `#expect`)
rather than XCTest. The Airbnb config enables rules that only apply to it —
`swiftTestingTestCaseNames`, `redundantSwiftTestingSuite`, `testSuiteAccessControl`,
`validateTestCases`, `noForceTryInTests`, `noForceUnwrapInTests`, `noGuardInTests`
— and every one of them is inert against XCTest code. XCTest is still required
for UI tests (`XCUIApplication`).
