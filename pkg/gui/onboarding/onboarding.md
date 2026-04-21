# Welcome to Lazyruin

#lazyruin-onboarding

## Tutorial

Welcome to `lazyruin`. Work through this checklist to learn the basics. If at any point you're done with this tutorial, type `:cleanup` to remove it from your vault.

During some parts of this tutorial, you will navigate away from this file. You can navigate back with `[` (or `:view history` and select the correct line) to come back to the exact line you were on. Or just click the note and hit `Enter` in the Notes panel.

### Intro

**What is lazyruin?** A terminal user interface (TUI) for the `ruin` notes CLI. Think of it like lazygit for your notes — it provides a keyboard-driven, interactive way to browse and organize your markdown notes.

**What is a vault?** A vault is a directory that holds all your notes as markdown files.

**The basic workflow:** Create notes with `#tags` and `[[wiki-style links]]` to connect them. Set up parent-child relationships to build hierarchies. Use Search and Pick to find specific content, and use Compose to view related notes together. Everything is stored as plain markdown files in your vault and is viewable in other programs.

### Basic Navigation
- [ ] Use `j`/`k` or arrow keys to move up and down inside a note or lists.
- [ ] Press `x` to mark a todo as complete. This also toggles a completed todo to be incomplete. Completed todos are moved to the bottom of the todo lists.
- [ ] `Esc` closes most dialogs. `q` quits `lazyruin`.
- [ ] Press `1`, `2`, `3` to focus Notes, Queries, Tags. Press the number again to switch tabs within a panel. Mouse clicks are also supported.
- [ ] From the side panel, press `Enter` to view a note. `Esc` to go back to the side panel.
- [ ] See all keyboard shortcuts with `?`.
- [ ] Use the command palette with `:`, which supports all keyboard shortcuts and a few additional commands.
- [ ] Use Quick Open to view a specific file with `<c-o>` or by typing `:` in the command palette.

### Your First Note
Now we will create your first note. For now, editing is basic. Use `/` to see formatting suggestions. Suggestions for `#tags` and `[[wiki-style links]]` are available as you type, press `Enter` or `Tab` to fill in a suggestion. Create a note that looks something like this:

```
# My First Note
#first
```

- [ ] `n` to make a new Note, `<c-s>` (ctrl-s) to save, `Esc` to cancel and close the dialog.

Now we'll make a second note, which will have two key ruin features: A parent and an inline tag. 

Make the content of the second note `- Remember the milk #remember`. Then type `>` to see the parent menu, select your first note.

These notes are now related and you can view them as a single note, when you'd like.

### Compose
- [ ] Navigate to your first note, either via the side-panel or Quick Open `<c-o>`.

You can now see it renders the child note in addition to your first note. This is due to the parent-child relationship we defined and `ruin compose`. Embeds `![[title]]` and dynamic embeds `![[pick: #followup]]` are also supported. See [advanced Compose documentation](https://github.com/donnellyk/ruin-note-cli/blob/main/docs/compose-advanced.md) for more info on syntax and capabilities (use `l` to highlight link with the cursor then `o` to open in your browser. You can also click the link on supported terminals).

- [ ] Bookmark this note with `b`. 

Bookmarking will make it available in the `Parent` side panel and display higher in `>` suggestions. Bookmark important files you think you'll come back to or reference frequently.

Within a composed note, `Enter` will jump you to the source child (or embedded) note or directly edit that within lazyruin with `e` or with your external $EDITOR with `E`.

### Pick & Search
- [ ] Search with `S`, type-ahead suggestions are available.
- [ ] Extract specific lines via Pick `p`. Pick `#remember` or the inline tag you added to your second note.
- [ ] Mark the line as `#done` with `D`.

Pick can extract specific context from a note, without you needing to read (be distracted) by the rest of the note. `#done` is a special reserved tag to mark an inline tag as complete. By default, all pick commands exclude lines with `#done` in results (ie. `#tag && !#done`). Done lines can be returned with `--any`.

Within a note, `#done` lines (and sections) are dimmed. You can toggle dimming or hide them entirely (along with overview options) via the view menu `v`.

Both Pick and Search support robust date filtering, see `ruin search --help` and `ruin pick --help` for more info.

### Today
- [ ] Open the command palette with `:` then type `today`. Select `Date: Today`.

This view shows all notes created or updated on the current date, along with any inline tags or todos annotated with today's date. In the New Note dialog, you can tag a line with `@YYYY-MM-DD` or common English words like `@today`, `@tomorrow`, `@next-week`.

You can view any day via the calendar `c`, contribution chart `C` or with command palette helpers like `:tomorrow`.

### Clean up
That's enough to get you started. See the [ruin-note-cli](https://github.com/donnellyk/ruin-note-cli) for more documentation.

Clean up this onboarding with `:cleanup` (or just delete it in the Notes panel with `d`). You can add it back at any time with `:add walkthrough`. 

If you want to remove the files you created during the tutorial, you can select them in the Notes panel and hit `d`.

Want to learn more? A more advanced tutorial is below.

### Advanced Features

#### Navigation
- [ ] In a list of cards (ie. Search or Pick results), `J`/`K` jump between cards. 
- [ ] `}`/`{` jumps to Next/Previous header.
- [ ] `l`/`L` jumps to Next/Previous link. `o` opens the link in your browser.
- [ ] `]`/`[` jumps Forward/Backward in history. You can view the full history with `:View History`.

#### Tags

A few additional tag formats are supported:

- `#tags-with-dashes`
- `#tags_with_underscores`
- `#nested/tags`
- `#tags with spaces#`

#### Parent

Child cards can, themselves, be parents to other cards, creating rich structures and hierarchies. To help with this, use `/` in the `>` parent suggestion to see and select a child card.

Example:
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

By default, children inherit their parent's tags. So, in the above example, a search for `#project_a` would return all three notes. To disable this behavior, set `tag_inheritance: false` in `~/.config/ruin/config.yml` or run `ruin config tag_inheritance false` (which sets it for you). See [configuration documentation](https://github.com/donnellyk/ruin-note-cli/blob/main/docs/configuration.md) for more options. There are also `lazyruin`-specific configuration options [here](https://github.com/donnellyk/lazyruin/blob/main/docs/configuration.md).

#### Embeds

In addition to parent-child relationships, you can also compose notes via the embed syntax. `![[title]]` embeds a specific notes (`![[title#header]]` to target a specific section). Dynamic queries like `![[search: ]]` and `![[pick: ]]` are also supported.

Example:
```
## Project A Followup
![[pick: #followup | filter=#project-a]]
```
will display as:

```
## Project A Followup

### Feb 23, 2026
- Simplify onboarding #idea #followup

### Feb 15, 2026
- Fix crash on startup when vault is empty #bug #followup
```

#### Links
A link is a special note type that is automatically tagged with `#link`. When a note is saved that meets the criteria of a link (the first line or first line after an H1 header is a valid URL), it's automatically tagged as `#link`. `<c-l>` presents a dedicated Link input that resolves a title and content from the URL automatically.

You can view all links with `:Browse Links` and further filter that list with `F`. They are also queryable in Search with `#link`.

By default, saving a new note via the New Note dialog that is _only_ a URL (and no other text, titles, tags, etc), the resolve step is shown. You can disable this behavior with `disable_bare_url_as_link: false` in `~/.config/lazyruin/config.yml`.

See [advanced Compose documentation](https://github.com/donnellyk/ruin-note-cli/blob/main/docs/compose-advanced.md) for more examples.
