# Welcome to LazyRuin

#lazyruin-onboarding

Work through this checklist to learn the basics. `[` / `]` navigate back / forward at any time.

## Punchlist
- Default to compose everywhere, but togglable (invert vh, basically)
- Remove `>>`, just rank the bookmarked parents higher. Two sections?
- Compose link
- Remove ! for now.
- Swap parent and query order.
- clicking into Preview should commit
- F isnt working?
- Add --all for pick
 
## Tutorial

### Basic Navigation
- [ ] Use `j` / `k` or arrow keys to move up and down inside a note or lists.
- [ ] Press `x` to toggle a todo on your current line.
- [ ] Press `1`, `2`, `3` to focus Notes, Queries, Tags. Press the number again to switch tabs within a panel. Mouse clicks are also supported.
- [ ] Press `Enter` to view a note, `Esc` to go back to the side panel.
- [ ] See all keyboard shortcuts with `?`.
- [ ] Use the command palette with `:`, which supports all keyboard shortcuts and a few additional commands.
- [ ] Quick Open to specific file with `<c-o>` or by typing `:` in the command palette.

### Your First Note
- [ ] `n` to make a new Note, `<c-s>` (ctrl-s) to save, `Esc` to cancel & close the dialog.`

Currently only basic editing exists. Use `/` to see formatting suggestions. Suggestions for `#tags` and `[[wiki-style links]]` are available as you type, press `Enter` or `Tab` to fill in a suggestion. Create a note that looks something like this:
```
# My First Note
#first
```

Make a second note, which will have two key Ruin features:
- Setting a parent
- An inline tag
```
- Remember the milk #remember
```

Before you save, type `>` to see the parent menu, select your first note. The H1 was automatically extracted as the title of the note, so you can reference it throughout Ruin with that title.

### Compose
- [ ] Go to your first note, either via the side-panel or Quick Open `<c-o>`

You can now see it renders the section note in addition to your first note. This is due to the parent-child relationship we defined and `ruin compose`. Embeds `![[title]]` and dynamic embeds `![[pick: #followup]]` are also supported. See (link)[link] for more info on syntax and capabilities (use `l` to highlight link then `o` to open in your browser. Or click it on supported browsers).

- [ ] Bookmark this note with `b`. 

Bookmarking will make it available in the `Parent` side panel & display higher in `>` suggestions. Bookmark important files you think you'll come back to or reference frequently.

Within a composed note, `Enter` will jump you to the source child (or embedded) note or directly edit that within lazyruin with `e` or with your external $EDITOR with `E`.

### Pick & Search
- [ ] Search with `S`, type-ahead suggestions are available 
- [ ] Extract specific lines via Pick `p`. Pick `#remember` or the inline tag you added to your second note.
- [ ] Mark the line as `#done` with `<c-d>`

Pick can extract specific context from a note, without you needing to read (be distracted) by the rest of the note. `#done` is a special reserved tag to mark an inline tag as complete. By default, all pick commands exclude lines with `#done` in results (ie. `#tag && !#done`). Done lines can be returned with `--any`.

Within a note, `#done` lines (and sections) are dimmed. You can toggle dimming or hide them entirely (along with over view options) via the view menu `v`.

Both Pick and Search support robust date filtering, see `ruin search --help` and `ruin pick --help` for more info.

### Today
- [ ] Open the command palette with `:` then type `today`. Select `Date: Today`

This view shows all notes created or updated on the current date, allow with any inline tags or todos annoted with today's date. In the New Note dialog, you can tag a line with `@YYYY-MM-DD` or common english words like `@today`, `@tomorrow`, `@next-week`.

You can view any day with via the calendar `c`, contribution chart `C` or with command palette helpers like `:tomorrow`

### Clean up
That's enough to get you started. See the [repo](link) for more documentation.

Clean up this onboarding with `:cleanup` (or just delete it in the Files panel with `d`). You can add it back at any time with `:add walkthrough`. 


### Basic Navigation
- [ ] In lists, press `j` / `k` or arrow keys to navigate. Scrolling with a mouse also works.
- [ ] Press `[` to go back. Press `]` to go forward.
- [ ] Press `n`. Type a note. `<c-s>` to save.
- [ ] In the new-note popup, try completions: `#` tags, `[[` wiki-links, `@` dates, `>` parents.
- [ ] Press `e` on a note for inline edit. `E` for `$EDITOR`.
- [ ] Press `S`. Search for anything. Try `sort:created` or `before:2025-01-01`.
- [ ] Press `p`. Pick by `#tag`. `<c-t>` for todo-only.
- [ ] Press `c`. Pick a date with arrow keys, `Enter` to view.
- [ ] Press `C`. Browse the contribution heatmap.
- [ ] Press `:`. Fuzzy-search any command.
- [ ] Press `<c-o>`. Jump to any note by title.
- [ ] Focus Tags (`3`). `Enter` on a tag to filter.
- [ ] In the preview, press `l` / `L` to cycle links; `o` to open.
- [ ] Press `<c-l>` to create a note from a URL.
- [ ] Try embeds in a note body: `![[Note Title]]` (static) or `![[pick:#tag]]` (dynamic).
- [ ] On a note, press `>` to set a parent.
- [ ] In Queries → Parents (`2`, `2`), press `b` on a parent to bookmark it, `Enter` to view as Compose.
- [ ] From Compose, press `Enter` on any line to jump to the source note.
- [ ] Press `i` to open the inbox.
- [ ] Press `v` in the preview to toggle display options.
- [ ] Press `?` to see all keybindings.
- [ ] When done: `:` → `cleanup` → `Enter`. Deletes this note and the `#lazyruin-onboarding` tag.

To bring this back: `:` → `add walkthrough`.
