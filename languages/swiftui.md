# SwiftUI Guidelines

Framework-level SwiftUI behavior: hosting/constraint crashes, layout,
state/observation, lists, typography, windows, and testing strategy. For
Swift-the-language, AppKit menu-bar apps, and packaging, see
[`swift.md`](swift.md).

This document adapts the ruleset from
[tqbf/swiftui-app](https://github.com/tqbf/swiftui-app)'s
`SWIFTUI-RULES.md`, generalized for project-independent use; examples below
are rewritten, not copied verbatim.

## 1. Animation & transitions — the silent constraint-engine crashes

### Never insert or remove a view with a `.transition` inside hosted SwiftUI

```swift
// WRONG — can trigger `_postWindowNeedsUpdateConstraintsUnlessPostingDisabled`
if isVisible {
    DetailPane()
        .transition(.move(edge: .trailing).combined(with: .opacity))
}

// CORRECT — always mounted, dimension-animated instead of presence-animated
DetailPane()
    .frame(width: isVisible ? paneWidth : 0)
    .clipped()
    .accessibilityHidden(!isVisible)
```

**Why.** When a SwiftUI view is hosted inside an
`_NSConstraintBasedLayoutHostingView` chain — a `NavigationSplitView`, a
sheet, a popover, any AppKit-hosted SwiftUI tree — layout runs in passes. A
transition that inserts or removes a view *during* an active pass calls
`setNeedsUpdateConstraints` from inside the update, which `NSWindow` guards
with an `NSInternalInconsistencyException`. This reproduces inconsistently
across Macs and OS minor versions; passing on one machine proves nothing.

**Fix.** Keep the view permanently in the tree and animate a *dimension*
(width, height, opacity) rather than the view's *presence*. `.clipped()`
hides the collapsed content. `withAnimation { flag.toggle() }` is safe as
long as the resulting change is dimensional, not structural (an `if
visible { … }` branch is structural and reintroduces the crash).

### Scope `TimelineView` to the smallest leaf that needs it

```swift
// WRONG — the whole row re-evaluates every tick
TimelineView(.periodic(from: .now, by: 1)) { ctx in
    HStack {
        Icon()
        Text(elapsed(ctx.date))
        Spacer()
        TrailingMenu()
    }
}

// CORRECT — only the digits redraw
HStack {
    Icon()
    TimelineView(.periodic(from: .now, by: 1)) { ctx in
        Text(elapsed(ctx.date))
    }
    Spacer()
    TrailingMenu()
}
```

Every tick re-evaluates the body of whatever's inside the `TimelineView`;
pulling it as tight as possible around the ticking content keeps redraws
cheap. The same lesson applies to `Timer.publish`/`onReceive`. For
"next event at X" style displays, prefer `.periodic(from: .now, by: 60)`
over a `Timer` — it auto-pauses off-screen and runs on the right scheduler.
Use 60s ticks for clock-time displays, 1s for elapsed counters, and avoid
sub-second ticks without a concrete reason.

## 2. Layout — frames, priorities, and design tokens

### Don't combine `.frame(maxWidth: .infinity)` with `.layoutPriority(N)`

The pair has surprising semantics: priority `N` requests *ideal* size first,
and `maxWidth: .infinity` makes "ideal" effectively infinite — siblings can
collapse to 0pt. If one view (a title) needs first claim on space, use a
trailing `Spacer(minLength:)` so the title sizes to content, or cap the
*other* sibling's `maxWidth`, or both.

### `.fixedSize(horizontal: true, vertical: false)` makes a view inflexible

It claims its ideal width regardless of the parent's proposal. Use it
deliberately for things that must never compress (a numeric counter, a
fixed-width chip). Don't put it on text that could legitimately truncate —
it will overflow the row instead of eliding when the parent shrinks.

### Every `Text` in a row needs an explicit line limit and truncation mode

```swift
Text(title)
    .lineLimit(1)
    .truncationMode(.tail)
```

Without `.lineLimit`, `Text` wraps to multiple lines instead of truncating.
Without `.truncationMode`, the default is leading truncation, which
surprises users ("…long title" instead of "long tit…"). On a narrow row,
skipping this is the difference between "ellipsis and done" and "wraps to
two lines and turns a chip into a tall pill that eats sibling space."

### Centralize layout metrics and typography in one `Theme` enum

```swift
enum Theme {
    // Metrics
    static let horizontalInset: CGFloat = 24
    static let rowMinHeight: CGFloat = 60

    // Semantic colors
    static let secondaryText = Color.secondary

    // Fonts — semantic styles only, never a hardcoded point size
    static let title = Font.title2.weight(.semibold)
    static let body = Font.body
    static let caption = Font.caption
}
```

Two rules keep a design system honest:

1. **Fonts are semantic, never hardcoded sizes.** `.font(.system(size: 11))`
   fights Dynamic Type and OS metric updates; `.font(.caption)` tracks them.
   If a control supplies its own typography (`Button` style, `Toggle`,
   `LabeledContent`), don't override it — you'll fight the system metric.
2. **Emphasis is weight; de-emphasis is color.** Headings get
   `.semibold`/`.bold`; quieter text gets `.secondary` — never a *lighter*
   weight, which reads as disabled rather than de-emphasized.

Don't sprinkle magic numbers or ad hoc font modifiers across files. The rule
of thumb: the third time you type the same literal, it's a metric — move it
into `Theme` and name it. This is also what keeps alignment from drifting
between panes because of an off-by-2 in one file.

### Prefer `Spacer` to push, over `frame(maxWidth: .infinity, alignment:)` to grow

Pushing with a `Spacer` is local and predictable. Growing with alignment
composes unpredictably with priority and parent size proposals. Default to
`Spacer` unless you have a specific reason not to.

### Floating a panel above its anchor needs measured height, not an alignment guess

```swift
// WRONG — renders ON the anchor and grows off-window; alignmentGuide(.top)
// does not give you "grow upward from here" for an overlay of unknown height
someField
    .overlay(alignment: .topLeading) {
        popupPanel
            .alignmentGuide(.top) { $0[.bottom] + 4 }
    }

// CORRECT — measure, then offset upward by the measured height
someField
    .background {
        GeometryReader { proxy in
            Color.clear.preference(key: PopupHeightKey.self, value: proxy.size.height)
        }
    }
    .overlay(alignment: .topLeading) {
        popupPanel
            .background {
                GeometryReader { proxy in
                    Color.clear.preference(key: PopupHeightKey.self, value: proxy.size.height)
                }
            }
            .offset(y: -(popupHeight + 4))
            .opacity(popupHeight > 0 ? 1 : 0)
    }
    .onPreferenceChange(PopupHeightKey.self) { popupHeight = $0 }
```

**Why.** There is no alignment-guide trick that reliably floats a view above
an anchor of unknown height — `alignmentGuide(.top)` repositions the panel's
own top edge, not "grow upward from the anchor." A naive attempt renders the
panel on top of the anchor and lets it grow downward, off-window.

**Fix.** Measure the panel's real height with a `GeometryReader` in
`.background` feeding a `PreferenceKey`, then `.offset(y: -(height + gap))`.
Gate visibility with `.opacity(height > 0 ? 1 : 0)` so there's no one-frame
flash at the wrong position before the height is known.

**Verify.** Don't trust a screenshot alone — an off-window panel can be
invisible in a screenshot too. Drive it with UI-test automation and assert
the popup's frame against the anchor's and the window's frames numerically.

## 3. State, observation, and frozen snapshots

### A cache built from a snapshot of an `@Observable` object goes stale silently

```swift
struct AvailableItem {
    var item: Item          // ← frozen snapshot!
    var resolved: Resolved
}

private(set) var available: [AvailableItem] = []   // cache

func setStatus(_ item: Item, to status: Status) {
    items[idx].status = status     // mutates the source
    persist(items[idx])
    Task { await rebuildAvailable() }   // REQUIRED — the cache does not know
}
```

**Why.** `@Observable` propagates "this property changed" through the direct
observation graph. It does not chase down copies embedded in other
observable properties. A cache derived from a snapshot has to be rebuilt
explicitly after every mutation that could affect it — an incremental patch
that tries to mirror each setter's side effects will eventually miss one. A
full rebuild from source is correct by construction; treat "smart"
incremental cache patching as a bug magnet, not an optimization.

### `@SceneStorage` defaults only apply on first scene initialization

Flipping a default from `true` to `false` does not migrate existing users —
their cached value sticks. If you're changing a default for a visibility/UX
reason, write the new value into storage explicitly as a migration step, or
bind the value to per-window ephemeral state instead, or move it to
versioned user preferences.

### `@Observable` identity is reference identity, not value identity

If an `@Observable` class is used in a SwiftUI `id:` or `ForEach` binding,
identify it by an immutable property (a UUID), never by the instance itself,
and don't rely on `==` meaning "value-equal."

### Read state at the latest possible moment

Don't capture a piece of shared state into a `let` at view construction time
and write it back later — a concurrent change in between gets clobbered on
save. Read fresh at the point of the action (a button handler, a save
trigger), mutate, then write.

## 4. Lists, scrolling, and rows

### `.swipeActions` only works inside `List`/`Form`

Use `.contextMenu` as the fallback for cards in a plain `ScrollView`. It's
fine to build both — users discover whichever is available for the
container they're in.

### `.listStyle(.inset)` + `.scrollContentBackground(.hidden)` is the modern macOS list look

Default `List` styling reads as a ported cross-platform widget. That pair
matches system app conventions (Mail, Reminders, Notes); pick one look and
apply it everywhere in the app.

### Row content: `.listRowSeparator(.hidden)` + explicit `.listRowInsets`

The system separator is heavier than most modern list designs want; let
spacing or card backgrounds carry the visual rhythm instead, and use insets
to give rows breathing room the system default doesn't provide.

### Build one generic row view, not a fork per screen

```swift
struct ItemRow<HoverActions: View, Trailing: View>: View {
    let title: String
    var subtitle: String?
    var isSelected: Bool = false
    @ViewBuilder var hoverActions: () -> HoverActions
    @ViewBuilder var trailing: () -> Trailing
}
```

A generic type with `EmptyView`-default overloads keeps the call site simple
while letting every screen share one row anatomy — changes to selection
tint, hover behavior, or spacing land once instead of N times.

### Hover-revealed actions: fade opacity, never insert-on-hover

```swift
hoverActions()
    .opacity(isHovered ? 1 : 0)
    .allowsHitTesting(isHovered)
    .animation(.easeOut(duration: 0.12), value: isHovered)
```

Inserting a view on hover reflows the row's width as the cursor crosses the
boundary, which reads as broken. Fading opacity keeps row geometry stable.

### On macOS, drive hover state yourself

`.hoverEffect()` is an iOS-only no-op on macOS. Use `.onHover { isHovered =
$0 }` and build your own background-tint priority (selected > next-up >
hovered > clear) behind the row content.

### `AsyncImage` cancels its fetch when its hosting row scrolls off

**Symptom.** Images in a `LazyVStack`/`List` load slowly, never retry after
a failure, and sometimes only complete after a manual scroll happens to
bring the row back into view.

**Why.** `AsyncImage` ties its network fetch to the hosting view's
lifetime via `.task`. In a lazily-recycled container, a row scrolling
off-screen — or the scroll view reflowing during an auto-scroll animation —
tears the view down and cancels the in-flight download before it finishes.

**Fix.** Own the fetch outside any one view's lifetime:

```swift
@MainActor
final class ImageCache {
    static let shared = ImageCache()
    private let cache = NSCache<NSURL, PlatformImage>()
    private var inFlight: [URL: Task<PlatformImage?, Never>] = [:]

    func image(for url: URL) async -> PlatformImage? {
        if let cached = cache.object(forKey: url as NSURL) { return cached }
        if let existing = inFlight[url] { return await existing.value }
        let task = Task<PlatformImage?, Never> { [weak self] in
            let data = try? await URLSession.shared.data(from: url).0
            let image = data.flatMap(PlatformImage.init)
            if let image { self?.cache.setObject(image, forKey: url as NSURL) }
            self?.inFlight[url] = nil
            return image   // failures are not cached — let a reappearing row retry
        }
        inFlight[url] = task
        return await task.value
    }
}
```

The fetch `Task` belongs to the cache, not the view: a cancelled *caller*
await (the view disappeared) does not cancel the underlying download — it
completes and populates the cache for whichever view asks next. Build a thin
`CachedAsyncImage` wrapper around this instead of the built-in `AsyncImage`
anywhere images live inside a lazily-recycled container.

### `ScrollViewReader.scrollTo` silently no-ops on an id that isn't rendered

**Symptom.** After prepending older items to a paginated list, an attempt to
re-anchor the scroll position at the previously-first item sometimes does
nothing — the viewport stays at the top of the grown content, which can
re-trigger a "load more" `onAppear` in a runaway loop.

**Why.** If the list applies any filtering, grouping, or collapsing between
the raw data model and what's actually rendered (e.g., collapsing a run of
minor events into one summary row keyed by the first event's id), the "raw"
oldest-item id may not correspond to any view actually present in the tree.
`scrollTo` targeting a missing id is a silent no-op, not an error.

**Fix.** Track two different ids: the pagination cursor (always the true
data-model boundary, so paging doesn't refetch the same page forever) and
the scroll anchor (always an id that is guaranteed to be rendered for
whatever the current display transformation is). Only pass the scroll
anchor to `proxy.scrollTo`. Also clear any pending anchor when the
underlying scroll container gets rebuilt (e.g., switching away from and
back to a paginated view before a fetch resolves) — an anchor computed
against the *old* content applied to the *new*, differently-anchored content
causes a visible jump.

### `ForEach` needs distinct identity per rendered occurrence, not per model object

**Symptom.** The same underlying item is meant to render in two places at
once (e.g., a "pinned" section and its normal section). Feeding both
`ForEach` loops the item's own stable id causes broken lazy-stack diffing —
lazy stacks require every rendered element in the identity space to be
unique, and the same id appearing twice violates that even across separate
`ForEach`s in the same container.

**Fix.** Identity is a rendering concern, not a data concern. Wrap the item
with its placement and derive the `id` from the composite:

```swift
struct RowOccurrence<Item: Identifiable>: Identifiable {
    enum Placement: Hashable { case pinned, section(SectionID) }
    struct ID: Hashable { let placement: Placement; let itemID: Item.ID }

    let item: Item
    let placement: Placement
    var id: ID { ID(placement: placement, itemID: item.id) }
}
```

Map every array through the occurrence wrapper before it reaches `ForEach`,
instead of iterating the model type directly wherever it might appear twice.

## 5. Text, typography, and dates

### Cache `DateFormatter`/`RelativeDateTimeFormatter` as `static let`

```swift
private static let relativeFormatter: DateFormatter = {
    let f = DateFormatter()
    f.dateStyle = .short
    f.timeStyle = .short
    f.doesRelativeDateFormatting = true
    return f
}()
```

Building a formatter inside `body` allocates one per render. On a
`TimelineView`-driven row, that's dozens of allocations per minute per row.

### Sanitize source strings before truncating them

Strip trailing periods, ellipses, or whitespace runs at the data layer.
Combining `.truncationMode(.tail)` with a string that already ends in `...`
produces both `...` *and* `…` in the same row — a bug that looks like a
SwiftUI rendering glitch but is actually upstream data hygiene.

### Build one component per repeated concept, parameterized by data

If the same concept (a status chip, a deadline badge) renders in three
different screens, it should be one SwiftUI `struct`, not three near-copies.
Compute any time-sensitive derived state (like "is this overdue") at the
leaf component itself, driven by a `TimelineView`, rather than at every call
site — callers should just pass the raw data.

### `foregroundStyle` resolves innermost-wins

```swift
// The whole line inherits `tint`; the trailing tag reasserts its own color.
HStack {
    Text(label)          // inherits tint from the container
    StatusTag(status)    // sets its own .foregroundStyle, wins locally
}
.foregroundStyle(tint)
```

The nearest `foregroundStyle` to a given `Text` wins, so a container can set
one tint for a whole line while a specific child overrides it locally — no
need to restructure the hierarchy to keep them apart. Corollary: if a child
should inherit the container's tint, don't set `.foregroundStyle(.primary)`
on it "to be explicit" — that silently opts it out of the ancestor tint.

## 6. Settings, windows, and toolbars

### Prefer `@Environment(\.openSettings)` over `SettingsLink` in nested containers

`SettingsLink` carries its own hosting-view shim that can trigger constraint
invalidation when nested inside `safeAreaInset`, a `List` inset, or other
already-hosted contexts. A plain `Button { openSettings() }` reaches the
same `Settings { }` scene without the hazard. Trade-off: `openSettings()`
can't pre-select a settings tab; if that matters, host `SettingsLink` at the
`NavigationSplitView` root or in a toolbar, not inside a `safeAreaInset`.

### Don't lead a `ToolbarItemGroup(placement: .primaryAction)` with `Spacer()`

The placement already trails its items; a leading `Spacer()` is redundant
and can confuse the toolbar's intrinsic-content-size computation badly
enough to crash on launch on some macOS versions. Just list the items.

### Keep toolbar items that read observable state shallow

A toolbar item reading a computed property off a shared store re-evaluates
on every redraw of that store. Either cache the computed value, or push the
read into a small dedicated `View` so SwiftUI's dependency tracking scopes
the redraw to just that item.

### Window/sheet scenes don't auto-dismiss when their driving state disappears

If a sheet or window is presented from `pendingFlow != nil`, SwiftUI can
restore that scene later (e.g. cold launch state restoration) without the
flag set, leaving a ghost window whose "context not found" branch never
self-dismisses. Make the host actively dismiss itself when its expected
state is absent, rather than assuming presentation and dismissal are always
symmetric.

### `scenePhase` tracks window visibility on macOS, not app activation

**Symptom.** Logic gated on `@Environment(\.scenePhase)` (e.g. "re-check a
live connection when the user returns to the app") rarely fires when the
user alt-tabs back to the app on macOS.

**Why.** On macOS, `scenePhase` reflects the window's own
visibility/occlusion, not the application's foreground/background state.
Switching via alt-tab, Mission Control, or Spaces routinely leaves a window
"active" by `scenePhase`'s accounting while the app was actually
backgrounded.

**Fix.** On macOS, observe `NSApplication.didBecomeActiveNotification`,
`didResignActiveNotification`, and `NSWorkspace.didWakeNotification`
directly via `.onReceive`, in addition to (not instead of) the
cross-platform `scenePhase` handler — iOS's `scenePhase` already tracks
foreground/background correctly and needs no extra observation.

## 7. Empty states, hover, and selection

### Every selectable destination needs a designed empty state

`ContentUnavailableView(title, systemImage:, description:) { actionButton }`
is the native idiom for "nothing selected yet" or "no results." One SF
Symbol, one clear description, one obvious action if there is one — never
ship blank canvas for a deselected or empty state.

### Background tint priority on a row should be explicit and ordered

```swift
private var backgroundStyle: AnyShapeStyle {
    if isSelected { return AnyShapeStyle(.accent.opacity(0.12)) }
    if isNext     { return AnyShapeStyle(.accent.opacity(0.06)) }
    if isHovered  { return AnyShapeStyle(Color.gray.opacity(0.08)) }
    return AnyShapeStyle(Color.clear)
}
```

Document the priority order in a comment. Make "the next thing to do" and
"the thing that's selected" visually distinct — same hue, different
intensity — so users can tell "recommended" from "currently open" at a
glance.

### Don't suppress the keyboard focus ring without a real reason

The default focus ring on a focused row is the correct affordance for
keyboard users. Calling `.focusEffectDisabled()` should be a deliberate,
justified choice, not a default reach because the ring "looked messy."

### Give every actionable row a right-click context menu on macOS

Wire menu items to the same methods the visible buttons call. Never include
a menu item whose action doesn't exist yet — a stub that does nothing is
worse than not offering the option.

## 8. Keyboard input

### `onKeyPress` arrow keys always carry implicit modifiers

**Symptom.** A key handler guarding `press.modifiers.isEmpty` to detect an
unmodified arrow-key press never fires for arrow keys at all.

**Why.** Arrow (and other function-row) keys report non-empty `modifiers`
even with no chord held — the platform sets implicit flags such as
`.function` and `.numericPad`. An `isEmpty` check rejects every arrow press,
not just chorded ones.

**Fix.** Check intersection with the real chord modifiers instead of
emptiness:

```swift
.onKeyPress(keys: [.upArrow, .downArrow]) { press in
    let chordModifiers: EventModifiers = [.command, .option, .control, .shift]
    guard press.modifiers.intersection(chordModifiers).isEmpty else {
        return .ignored   // let a real chord (e.g. ⌥↑) fall through to app-level shortcuts
    }
    return handleArrow(up: press.key == .upArrow) ? .handled : .ignored
}
```

## 9. Live parsing and preview UI

These rules apply to any UI that shows a live-parsed preview of user input —
a quick-add capture field, a search bar with structured query parsing, a form
field with date or quantity detection.

### Never present a parse result as read-only

Users need an escape hatch. Make the parsed chip clickable (opens an editor)
and clearable (a dismiss button removes it). Natural-language parsing fails
too often to be a one-way door — a "task title tomorrow at 3pm" input that
mis-parses the date needs a way to fix it without retyping the whole thing.

### Surface confidence visually

If the parser can report low confidence, propagate it to the UI: tint the
chip differently (e.g. yellow instead of the confident color) and swap its
icon (e.g. to a question mark). A low-confidence chip says "this is my best
guess — verify before accepting," rather than presenting a guess as fact.

### Strip the matched phrase from the remaining text preview

If the parser matched "tomorrow at 3pm" out of the input "renew passport
tomorrow at 3pm", the preview title should read "renew passport" — not the
raw input alongside the chip. Showing both leaves the user unsure whether
the literal words "tomorrow at 3pm" will also be saved as text.

### Tooltip the chip with the absolute value

Relative displays ("Tomorrow at 3:00 PM") are scannable but ambiguous across
timezones and week boundaries. `.help(absoluteFormatter.string(from: date))`
(or a long-press tooltip on iOS) gives a verification path without
cluttering the chip itself.

### Gate dependent UI on the parse outcome

```swift
// A follow-up control that only makes sense once a date was actually parsed.
if let dueDate {
    ReminderOffsetPicker(around: dueDate)
}
```

Don't render a dependent control unconditionally and let the user wonder why
their selection appears to do nothing — gate it on the parse result being
non-nil (or otherwise present).

### Surface Esc and Return explicitly

The standard chord for "dismiss" is Esc; for "commit" is Return. Render
small keycap affordances in the overlay's footer ("esc to dismiss," "return
to save") — don't rely on the user already knowing the muscle memory.

## 10. Sandboxing and entitlements

### Default sandboxed macOS apps to explicit read/write entitlements, one per capability

A sandboxed app (`ENABLE_APP_SANDBOX=YES`) needs an explicit entitlement for
each system-resource capability it uses — file access, network, camera,
and so on — declared in its `.entitlements` file. Treat sandboxing as the
default posture for a distributable macOS app, and add each entitlement
only when a concrete feature needs it, so the capability surface stays
auditable.

### `.fileImporter` crashes without the user-selected read entitlement

**Symptom.** A SwiftUI `.fileImporter` (backed by `NSOpenPanel`) hard-crashes
with `EXC_BREAKPOINT` in a sandboxed app, with a console message about a
missing "User Selected File Read" sandbox entitlement.

**Why.** The sandbox denies the file-read grant `NSOpenPanel` needs to
return a security-scoped URL, and SwiftUI does not degrade gracefully — it
traps instead of surfacing a catchable error.

**Fix.** Add `com.apple.security.files.user-selected.read-only` (or
`.read-write` if the picker needs to write back) to the app's entitlements
file. Drag & drop (`onDrop`) does **not** need this entitlement — dropped
data arrives over the drag pasteboard, a different and already-granted
sandbox gate.

**Gotcha.** Entitlement changes require a full rebuild and re-sign; an
already-running process keeps its old sandbox container and will not pick up
a newly added entitlement.

## 11. Xcode Previews hosting quirks

### A sidebar `List(selection:)` with `Section`s can crash the Previews host, not the app

**Symptom.** A `List(selection:)` with `.listStyle(.sidebar)` and `Section`s
crashes the Xcode Previews process specifically — a fatal error deep in
private `NSOutlineView`-backed list internals during outline expansion —
while the same view runs fine in the actual built app.

**Why.** The Previews host's key-view-loop recalculation force-expands the
`NSOutlineView` backing a sidebar-style `List`, and an internal
selection-update diff can assert. This can fire even when the bound
selection value is `nil` — the mere presence of a selection binding is
enough to trigger it. A `List` with no selection binding renders fine.

**Fix.** Give the Previews entry point a selection-less `List` and give the
real app entry point `List(selection:)`, sharing the row content via one
`@ViewBuilder`:

```swift
var body: some View {
    if ProcessInfo.isPreviewOrTest {
        List { sidebarContent }
    } else {
        List(selection: $selection) { sidebarContent }
    }
}
```

Apply the same guard to anything else that needs a scene Previews doesn't
provide (e.g. `SettingsLink` needs a `Settings` scene). If a preview seeds
its model asynchronously, prefer a synchronous fixture-hydration path over
letting the real async load run inside Previews — an empty-to-populated
transition mid-render is a second, independent way into the same crash.

**Debugging tip.** When a crash report names a private AppKit/SwiftUI
symbol, treat your first theory as provisional. Read the newest crash log
from `~/Library/Logs/DiagnosticReports/`, parse it (most are JSON after a
short text header), and inspect the crashed thread's symbol stack directly
— that's usually the only way to confirm which UI construct triggered a
crash inside framework-private code.

## 12. Swift Charts

### A second series needs its own `series:` identity, or marks merge into one path

```swift
// WRONG — both LineMarks default to the SAME series; Charts connects every
// point into one line and one resolved style wins (drawing order decides).
Chart {
    ForEach(indoor)  { LineMark(x: .value("Time", $0.time), y: .value("Value", $0.value)) }
    ForEach(outdoor) { LineMark(x: .value("Time", $0.time), y: .value("Value", $0.value)) }
}

// CORRECT — distinct series identity keeps them as two independently-styled lines
Chart {
    ForEach(indoor) {
        LineMark(x: .value("Time", $0.time), y: .value("Value", $0.value),
                 series: .value("Series", "Indoor"))
    }
    ForEach(outdoor) {
        LineMark(x: .value("Time", $0.time), y: .value("Value", $0.value),
                 series: .value("Series", "Outdoor"))
    }
}
```

**Why.** `LineMark`s are grouped into a connected path by their `series:`
value. With none specified, every line in the chart shares the same default
series, so Charts joins all the points into one path and one resolved style
wins. Discrete marks (`RuleMark`, `BarMark`) don't need this — color alone
distinguishes them.

### The x-domain is the union of everything plotted — clip overlays to match

**Symptom.** A chart meant to show one month of daily data suddenly spans a
full year once a second, longer-running series is overlaid: the short
series gets squeezed into a sliver at one edge while the long one sprawls
across the full width.

**Why.** With no explicit `chartXScale`, Charts derives the x-domain from
the union of *all* plotted data and zooms out to fit the longest series.

**Fix.** Clip every overlay series to the same window as the primary series
before plotting (a matching cutoff in the query or filter), or set
`chartXScale(domain:)` explicitly. Don't trust auto-domain once two series
disagree about how much history they carry.

## 13. Concurrency (Swift 6 strict)

### A `@MainActor` type adopting a delegate protocol needs an isolated conformance

```swift
// WRONG — "conformance … crosses into main actor-isolated code and can cause
// data races." The protocol's requirements are nonisolated; the impls aren't.
@MainActor final class LocationTracker: NSObject, CLLocationManagerDelegate { … }

// CORRECT — isolate the conformance itself (Swift 6 isolated conformances)
@MainActor final class LocationTracker: NSObject, @MainActor CLLocationManagerDelegate { … }
```

**Why.** Delegate protocols like `CLLocationManagerDelegate` declare their
requirements as `nonisolated` by default, since the compiler can't assume
the callbacks run on any particular actor in general — but a `@MainActor`
conforming type's implementations are, of course, isolated to the main
actor, and that mismatch is what the compiler flags. When the callbacks
genuinely only ever fire on the main thread — true whenever the delegating
object (e.g. a `CLLocationManager`) was itself created on the main actor —
annotating the conformance is the correct fix, and the compiler will even
suggest exactly this. The alternative — marking each method `nonisolated`
and hopping with `MainActor.assumeIsolated` inside every callback — is more
ceremony for no real gain in that situation. Reach for the `nonisolated` +
`assumeIsolated` pattern only when the delegate can genuinely fire off the
main actor and the isolated-conformance shortcut would be a lie.

## 14. Testing strategy

### The compile gate is necessary but not sufficient

`swift build` proves the code parses; `swift test` proves model logic works.
Neither proves the app launches, lays out correctly, or survives a clean
build. Add a "launch and stay alive" gate: run the built app, dismiss the
window, perform one user action, confirm the process is still alive. Most
constraint-engine crashes fire within the first display cycle.

### Test on the lowest-spec OS version the app supports

Constraint-engine strictness varies across OS minor versions — a pattern
that works on the newest release can throw on an older supported one. If
the deployment floor is N versions back, include a machine or CI runner on
that floor in the gate.

### Visual gates beat unit tests for layout and styling changes

If a row's anatomy, a chip's tint policy, or an empty state's copy changed,
the right gate is a screenshot, not a forced unit test. SwiftUI snapshot
tests are brittle enough that the cost-to-value ratio rarely pencils out.

### Write a regression test for every fix, even the "obvious" one-liners

A one-line fix (e.g., adding a missing cache-rebuild call) deserves a test
that locks in the scenario that broke without it. The fix is easy to
re-break in a future refactor if nothing guards it.

### A crash naming a private framework symbol makes your first fix provisional

Reason from the stack shape (a recursive "container needs subview update"
pattern usually means "a view was structurally modified mid-update"),
enumerate the likely suspects in the recent diff, fix the most likely one,
ship, and treat it as a two-shot fix — ask for a retest rather than assuming
the first attempt closed it.

## 15. Keeping a project history

A few lightweight habits pay for themselves on any SwiftUI project with real
churn:

- **Keep a running log.** Whatever it's called — `CHANGELOG.md`, a dated
  journal — append every meaningful change with what was done, what was
  learned, and what was surprising. Future readers, including yourself
  months later, reach for this when an unfamiliar bug looks familiar.
- **Keep a "things that bit us" file.** A separate file for hard-won pattern
  lessons — the constraint-crash pattern that took an afternoon to diagnose,
  the workaround that looked unnecessary until it wasn't. One entry per
  pattern, not folded into commit messages where it gets lost. This document
  is that file for SwiftUI framework behavior in general; a specific project
  should keep its own for lessons that don't generalize past its own
  codebase.
- **Commit messages explain the why, not the what.** The diff already shows
  what changed. The reader six months from now needs the failure mode that
  motivated the change, the alternatives considered, and the constraint that
  ruled them out.

For broader process guidance not specific to SwiftUI — batching work, review
cadence, general project management — see `project_tech_stack.md` at this
submodule's root; this section only covers the SwiftUI-flavored habits
above and isn't meant to duplicate or override that document.
