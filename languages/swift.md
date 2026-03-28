# Swift Guidelines

## macOS Menu Bar (Tray) Applications

### Entry Point — Use SwiftUI @main with @NSApplicationDelegateAdaptor

**Critical**: Always use the SwiftUI `@main` App lifecycle with `@NSApplicationDelegateAdaptor` for menu bar apps. Do NOT use the manual `NSApplication.shared.run()` pattern — it causes `NSPopover` positioning bugs where the popover appears detached from the status bar icon.

```swift
// CORRECT — popover anchors properly to status bar icon
@main
struct MyApp: App {
    @NSApplicationDelegateAdaptor(AppDelegate.self) var appDelegate

    var body: some Scene {
        Settings {
            EmptyView()
        }
    }
}
```

```swift
// WRONG — causes popover misplacement
// let app = NSApplication.shared
// app.setActivationPolicy(.accessory)
// let delegate = AppDelegate()
// app.delegate = delegate
// app.run()
```

### AppDelegate Pattern

The `AppDelegate` creates the status bar item and popover. Key points:

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

5. **Launch via `open -W`** in launchd plists — if the binary is launched directly (`Contents/MacOS/MyApp`), macOS doesn't associate the process with the `.app` bundle, so the bundle icon won't appear in notifications:

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

### Platform & Package Manager

- Use Swift Package Manager (SPM) with `Package.swift`
- Target `macOS(.v13)` or later for full SwiftUI support
- No extra dependencies needed for menu bar apps — SwiftUI and AppKit are system frameworks
- For bundled resources (icons), use `.process("Resources")` in the SPM target and access via `Bundle.module`
