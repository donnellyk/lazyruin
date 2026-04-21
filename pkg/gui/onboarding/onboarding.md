# Welcome to Lazyruin

#lazyruin-onboarding

## Tutorial

Welcome to `lazyruin`. Work through this checklist to learn the basics. If at any one you are done, type `:cleanup` to remove this tutorial from your vault.

Parts of the tutorial will have to navigate away from this file. You can navigate back with `[` (or `:view history` and select the correct line) to come back to the exact line you were on. Or just click the note and hit `Enter` in the Notes panel.

### Basic Navigation
- [ ] Use `j` / `k` or arrow keys to move up and down inside a note or lists.
- [ ] Press `x` to mark a todo as complete. This also toggles a completed todo to be incomplete. Completed todos are moved to the bottom of the todo lists.
- [ ] `Esc` closes most dialogs. `q` quits `lazyruin`.
- [ ] Press `1`, `2`, `3` to focus Notes, Queries, Tags. Press the number again to switch tabs within a panel. Mouse clicks are also supported.
- [ ] From the side panel, press `Enter` to view a note. `Esc` to go back to the side panel.
- [ ] See all keyboard shortcuts with `?`.
- [ ] Use the command palette with `:`, which supports all keyboard shortcuts and a few additional commands.
- [ ] Use Quick Open to specific file with `<c-o>` or by typing `:` in the command palette.

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
- [ ] Go to your first note, either via the side-panel or Quick Open `<c-o>`.

You can now see it renders the section note in addition to your first note. This is due to the parent-child relationship we defined and `ruin compose`. Embeds `![[title]]` and dynamic embeds `![[pick: #followup]]` are also supported. See [advanced Compose documentation](https://github.com/donnellyk/ruin-note-cli/blob/main/docs/compose-advanced.md) for more info on syntax and capabilities (use `l` to highlight link then `o` to open in your browser. Or click it on supported browsers).

- [ ] Bookmark this note with `b`. 

Bookmarking will make it available in the `Parent` side panel & display higher in `>` suggestions. Bookmark important files you think you'll come back to or reference frequently.

Within a composed note, `Enter` will jump you to the source child (or embedded) note or directly edit that within lazyruin with `e` or with your external $EDITOR with `E`.

### Pick & Search
- [ ] Search with `S`, type-ahead suggestions are available.
- [ ] Extract specific lines via Pick `p`. Pick `#remember` or the inline tag you added to your second note.
- [ ] Mark the line as `#done` with `D`.

Pick can extract specific context from a note, without you needing to read (be distracted) by the rest of the note. `#done` is a special reserved tag to mark an inline tag as complete. By default, all pick commands exclude lines with `#done` in results (ie. `#tag && !#done`). Done lines can be returned with `--any`.

Within a note, `#done` lines (and sections) are dimmed. You can toggle dimming or hide them entirely (along with over view options) via the view menu `v`.

Both Pick and Search support robust date filtering, see `ruin search --help` and `ruin pick --help` for more info.

### Today
- [ ] Open the command palette with `:` then type `today`. Select `Date: Today`.

This view shows all notes created or updated on the current date, along with any inline tags or todos annoted with today's date. In the New Note dialog, you can tag a line with `@YYYY-MM-DD` or common english words like `@today`, `@tomorrow`, `@next-week`.

You can view any day with via the calendar `c`, contribution chart `C` or with command palette helpers like `:tomorrow`

### Clean up
That's enough to get you started. See the [repo](link) for more documentation.

Clean up this onboarding with `:cleanup` (or just delete it in the Files panel with `d`). You can add it back at any time with `:add walkthrough`. 

If you want to remove the files you created during the tutorial, you can select them in the Notes panel and hit `d`.

Want to learn more? A more advanced tutorial is below.

### Advanced Features

#### Navigation
- [ ] In a list of cards (ie. Search or Pick results), `J` / `K` jump between cards. 
- [ ] `}` / `{` jumps to Next / Previous header.
- [ ] `l` / `L` jumps to Next / Previous link. `o` opens the link in your browser.
- [ ] `]` / `[` jumps Forward and Backward in history. You can view the full history with `:View History`.

#### Tags
A few additional tag formats are supported
- `#tags-with-dashes`
- `#tags_with_underscores`
- `#nested/tags`
- `#tags with spaces#` 

#### Parent
Child cards can, themselves, be parents to other cards, creating rich structures and hierarchies. To help with this, use `/` in the `>` parent suggestion to see & select a child card. 

Example
```
# Project A

#project_a, #2026/q2
```

```
## Updates
>Project A
```

```
# Apr 21, 2026
Met with #bob, need to ask him about release plan #followup
>Project A/Updates
```

Would be displayed, with `lazyruin` as

```
# Project A

## Updates

### Apr 21, 2026
Met with #bob, need to ask him about release plan #followup
```

By default, children inherit their parent's tags. So, in the above example, a search for `#project_a` would return all three notes. To disable this behavior, set `tag_inheritance: false` in `~/.config/ruin/config.yml` or run `ruin config tag_inheritance false` (which sets it for you). See [configuration documentation](https://github.com/donnellyk/ruin-note-cli/blob/main/docs/configuration.md) for more options. There are also `lazyruin` specific configuration options [here](https://github.com/donnellyk/lazyruin/blob/main/docs/configuration.md).

#### Embeds
In addition to parent-child relationships, you can also compose notes via the embed syntax. `![[title]]` embeds a specific notes (`![[title#header]]` to target a specific section). Dynamic queries like `![[search: ]]` and `![[pick: ]]` are also supported.

Example
```
## Project A Followup
![[pick: #followup | filter=#project-a]]
```
will display as 
```
## Project A Followup

### Feb 23, 2026
- Simplify onboarding #idea #followup

### Feb 15, 2026
- Fix crash on startup when vault is empty #bug #followup
```

See [advanced Compose documentation](https://github.com/donnellyk/ruin-note-cli/blob/main/docs/compose-advanced.md) for more examples
