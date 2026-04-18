package gui

import (
	"strings"

	"github.com/donnellyk/lazyruin/pkg/gui/context"
	"github.com/donnellyk/lazyruin/pkg/gui/controllers"
	helperspkg "github.com/donnellyk/lazyruin/pkg/gui/helpers"
	"github.com/donnellyk/lazyruin/pkg/gui/types"
	"github.com/donnellyk/lazyruin/pkg/models"

	"github.com/jesseduffield/gocui"
)

// setupNotesContext initializes the "notes" and NotesController.
func (gui *Gui) setupNotesContext() {
	notesCtx := context.NewNotesContext(gui.RenderNotes, func() { gui.helpers.Preview().UpdatePreviewForNotes() })
	gui.contexts.Notes = notesCtx
	gui.contextMgr.Register(notesCtx)

	notesCtx.AddOnFocusFn(func(_ types.OnFocusOpts) {
		gui.RefreshNotes(true)
		gui.helpers.Preview().UpdatePreviewForNotes()
	})

	gui.notesController = controllers.NewNotesController(controllers.NotesControllerOpts{
		Common:     gui.controllerCommon,
		GetContext: func() *context.NotesContext { return gui.contexts.Notes },
		OnShowInfo: func(_ *models.Note) error {
			return gui.helpers.PreviewInfo().ShowInfoDialog()
		},
	})

	controllers.AttachController(gui.notesController)
}

// setupTagsContext initializes the new "tags" and TagsController.
func (gui *Gui) setupTagsContext() {
	tagsCtx := context.NewTagsContext(gui.RenderTags, func() { gui.helpers.Tags().UpdatePreviewForTags() })

	gui.contexts.Tags = tagsCtx
	gui.contextMgr.Register(tagsCtx)

	tagsCtx.AddOnFocusFn(func(_ types.OnFocusOpts) {
		gui.RefreshTags(true)
		gui.helpers.Tags().UpdatePreviewForTags()
	})

	gui.tagsController = controllers.NewTagsController(controllers.TagsControllerOpts{
		Common:     gui.controllerCommon,
		GetContext: func() *context.TagsContext { return gui.contexts.Tags },
	})

	controllers.AttachController(gui.tagsController)
}

// setupQueriesContext initializes the "queries" and QueriesController.
func (gui *Gui) setupQueriesContext() {
	queriesCtx := context.NewQueriesContext(
		gui.RenderQueries, func() { gui.helpers.Queries().UpdatePreviewForQueries() },
		gui.RenderQueries, func() { gui.helpers.Queries().UpdatePreviewForParents() },
	)
	gui.contexts.Queries = queriesCtx
	gui.contextMgr.Register(queriesCtx)

	queriesCtx.AddOnFocusFn(func(_ types.OnFocusOpts) {
		if gui.contexts.Queries.CurrentTab == context.QueriesTabParents {
			gui.RefreshParents(true)
			gui.helpers.Queries().UpdatePreviewForParents()
		} else {
			gui.RefreshQueries(true)
			gui.helpers.Queries().UpdatePreviewForQueries()
		}
	})

	gui.queriesController = controllers.NewQueriesController(controllers.QueriesControllerOpts{
		Common:     gui.controllerCommon,
		GetContext: func() *context.QueriesContext { return gui.contexts.Queries },
	})

	controllers.AttachController(gui.queriesController)
}

// registerPreviewContext is a helper that handles the common registration
// boilerplate for preview contexts: store via setter, register with contextMgr,
// create a controller via ctrlFactory, and attach it. The onFocus callback
// is added to the context before the controller is attached.
func registerPreviewContext[C types.Context](
	gui *Gui,
	ctx C,
	setter func(C),
	onFocus func(types.OnFocusOpts),
	ctrlFactory func() types.IController,
) {
	setter(ctx)
	gui.contextMgr.Register(ctx)
	ctx.AddOnFocusFn(onFocus)
	ctrl := ctrlFactory()
	controllers.AttachController(ctrl)
}

// setupPreviewContext initializes the four preview contexts (cardList,
// pickResults, compose, datePreview). Navigation history is managed by the
// NavigationManager owned by the Navigator helper.
func (gui *Gui) setupPreviewContext() {
	gui.contexts.ActivePreviewKey = "cardList"

	onPreviewFocus := func(_ types.OnFocusOpts) { gui.RenderPreview() }

	registerPreviewContext(gui, context.NewCardListContext(),
		func(ctx *context.CardListContext) { gui.contexts.CardList = ctx },
		onPreviewFocus,
		func() types.IController {
			return controllers.NewCardListController(gui.controllerCommon,
				func() *context.CardListContext { return gui.contexts.CardList })
		},
	)

	registerPreviewContext(gui, context.NewPickResultsContext(),
		func(ctx *context.PickResultsContext) { gui.contexts.PickResults = ctx },
		onPreviewFocus,
		func() types.IController {
			return controllers.NewPickResultsController(gui.controllerCommon,
				func() *context.PickResultsContext { return gui.contexts.PickResults })
		},
	)

	registerPreviewContext(gui, context.NewComposeContext(),
		func(ctx *context.ComposeContext) { gui.contexts.Compose = ctx },
		onPreviewFocus,
		func() types.IController {
			return controllers.NewComposeController(gui.controllerCommon,
				func() *context.ComposeContext { return gui.contexts.Compose })
		},
	)

	registerPreviewContext(gui, context.NewDatePreviewContext(),
		func(ctx *context.DatePreviewContext) { gui.contexts.DatePreview = ctx },
		onPreviewFocus,
		func() types.IController {
			return controllers.NewDatePreviewController(gui.controllerCommon,
				func() *context.DatePreviewContext { return gui.contexts.DatePreview })
		},
	)

	gui.seedPreviewDisplayStateFromConfig()
}

// seedPreviewDisplayStateFromConfig applies persisted view options to every
// preview context after construction. Call after all preview contexts are
// registered; safe with nil config (no-op).
func (gui *Gui) seedPreviewDisplayStateFromConfig() {
	if gui.config == nil {
		return
	}
	hide := gui.config.ViewOptions.HideDone
	for _, ctx := range []context.IPreviewContext{
		gui.contexts.CardList,
		gui.contexts.PickResults,
		gui.contexts.Compose,
		gui.contexts.DatePreview,
	} {
		if ctx == nil {
			continue
		}
		ctx.DisplayState().HideDone = hide
	}
}

// registerPopupContext is a helper that handles the common registration
// boilerplate for popup contexts: store via setter, register with contextMgr,
// create a PopupController with the given bindings, and attach it.
func registerPopupContext[C types.Context](gui *Gui, ctx C, setter func(C), bindings []*types.Binding) {
	setter(ctx)
	gui.contextMgr.Register(ctx)
	ctrl := controllers.NewPopupController(func() C { return ctx }, bindings)
	controllers.AttachController(ctrl)
}

// setupSearchContext initializes the "search" and its popup controller.
func (gui *Gui) setupSearchContext() {
	searchState := func() *types.CompletionState { return gui.contexts.Search.Completion }
	searchHelper := func() *helperspkg.SearchHelper { return gui.helpers.Search() }

	registerPopupContext(gui, context.NewSearchContext(),
		func(ctx *context.SearchContext) { gui.contexts.Search = ctx },
		[]*types.Binding{
			{Key: gocui.KeyEnter, Description: "Search", Handler: func() error {
				return gui.completionEnter(searchState, gui.searchOrFilterTriggers, func(g *gocui.Gui, v *gocui.View) error {
					raw := strings.TrimSpace(v.TextArea.GetUnwrappedContent())
					ctx := gui.contexts.Search
					if ctx.InFilterMode() {
						err := ctx.OnFilterSubmit(raw)
						ctx.ClearFilterMode()
						searchHelper().CancelSearch()
						return err
					}
					if !searchHelper().ExecuteSearch(raw) {
						searchHelper().CancelSearch()
					}
					return nil
				})(gui.g, gui.views.Search)
			}},
			{Key: gocui.KeyEsc, Description: "Cancel", Handler: func() error {
				return gui.completionEsc(searchState, func(g *gocui.Gui, v *gocui.View) error {
					searchHelper().CancelSearch()
					return nil
				})(gui.g, gui.views.Search)
			}},
			{Key: gocui.KeyTab, Description: "Complete", Handler: func() error {
				return gui.completionTab(searchState, gui.searchOrFilterTriggers)(gui.g, gui.views.Search)
			}},
		},
	)
}

// setupCaptureContext initializes the "capture" and its popup controller.
func (gui *Gui) setupCaptureContext() {
	registerPopupContext(gui, context.NewCaptureContext(),
		func(ctx *context.CaptureContext) { gui.contexts.Capture = ctx },
		[]*types.Binding{
			{Key: gocui.KeyCtrlS, Description: "Save", Handler: func() error {
				content := strings.TrimSpace(gui.views.Capture.TextArea.GetUnwrappedContent())
				if gui.contexts.Capture.LinkURL != "" {
					return gui.helpers.Link().SubmitLinkCapture(content, gui.QuickLink)
				}
				return gui.helpers.Capture().SubmitCapture(content, gui.QuickCapture)
			}},
			{Key: gocui.KeyEsc, Description: "Cancel", Handler: func() error {
				return gui.helpers.Capture().CancelCapture(gui.QuickCapture || gui.QuickLink)
			}},
			{Key: gocui.KeyTab, Handler: func() error { return gui.captureTab(gui.g, gui.views.Capture) }},
			{Key: gocui.KeyCtrlJ, Description: "Jot to inbox", Handler: func() error {
				return gui.helpers.Inbox().OpenInboxInput()
			}},
			{Key: gocui.KeyCtrlO, Description: "Insert inbox item", Handler: func() error {
				return gui.helpers.Inbox().OpenBrowserForInsert(func(text string) {
					if gui.views.Capture != nil {
						gui.views.Capture.TextArea.TypeString(text)
					}
				})
			}},
		},
	)
}

// setupPickContext initializes the "pick" and its popup controller.
func (gui *Gui) setupPickContext() {
	pickState := func() *types.CompletionState { return gui.contexts.Pick.Completion }

	registerPopupContext(gui, context.NewPickContext(),
		func(ctx *context.PickContext) { gui.contexts.Pick = ctx },
		[]*types.Binding{
			{Key: gocui.KeyEnter, Description: "Pick", Handler: func() error {
				executePick := func(g *gocui.Gui, v *gocui.View) error {
					raw := strings.TrimSpace(v.TextArea.GetUnwrappedContent())
					return gui.helpers.Pick().ExecutePick(raw)
				}
				return gui.completionEnter(pickState, gui.pickTriggers, executePick)(gui.g, gui.views.Pick)
			}},
			{Key: gocui.KeyEsc, Description: "Cancel", Handler: func() error {
				cancelPick := func(g *gocui.Gui, v *gocui.View) error {
					return gui.helpers.Pick().CancelPick()
				}
				return gui.completionEsc(pickState, cancelPick)(gui.g, gui.views.Pick)
			}},
			{Key: gocui.KeyTab, Description: "Complete", Handler: func() error {
				return gui.completionTab(pickState, gui.pickTriggers)(gui.g, gui.views.Pick)
			}},
			{Key: gocui.KeyCtrlA, Description: "--any", Handler: func() error {
				gui.helpers.Pick().TogglePickAny()
				if gui.views.Pick != nil {
					gui.views.Pick.Footer = gui.pickFooter()
				}
				return nil
			}},
			{Key: gocui.KeyCtrlT, Description: "--todo", Handler: func() error {
				gui.helpers.Pick().TogglePickTodo()
				if gui.views.Pick != nil {
					gui.views.Pick.Footer = gui.pickFooter()
				}
				return nil
			}},
			{Key: gocui.KeyCtrlL, Description: "--all-tags", Handler: func() error {
				gui.helpers.Pick().TogglePickAllTags()
				if gui.views.Pick != nil {
					gui.views.Pick.Footer = gui.pickFooter()
				}
				return nil
			}},
		},
	)
}

// setupInputPopupContext initializes the InputPopupContext and its popup controller.
func (gui *Gui) setupInputPopupContext() {
	registerPopupContext(gui, context.NewInputPopupContext(),
		func(ctx *context.InputPopupContext) { gui.contexts.InputPopup = ctx },
		[]*types.Binding{
			{Key: gocui.KeyEnter, Handler: func() error {
				ctx := gui.contexts.InputPopup
				if ctx.Config != nil && ctx.Config.Locked {
					return nil
				}
				v, _ := gui.g.View(InputPopupView)
				raw := strings.TrimSpace(v.TextArea.GetUnwrappedContent())
				state := ctx.Completion
				var item *types.CompletionItem
				if state.Active && len(state.Items) > 0 {
					selected := state.Items[state.SelectedIndex]
					item = &selected
				}
				return gui.helpers.InputPopup().HandleEnter(raw, item)
			}},
			{Key: gocui.KeyEsc, Handler: func() error {
				ctx := gui.contexts.InputPopup
				if ctx.Config != nil && ctx.Config.Locked {
					if ctx.Config.OnCancel != nil {
						ctx.Config.OnCancel()
					}
					gui.helpers.InputPopup().CloseInputPopup()
					if gui.QuickLink {
						return gocui.ErrQuit
					}
					return nil
				}
				completionWasActive := ctx.Completion.Active
				if err := gui.helpers.InputPopup().HandleEsc(); err != nil {
					return err
				}
				if !completionWasActive && gui.QuickLink {
					return gocui.ErrQuit
				}
				return nil
			}},
			{Key: gocui.KeyTab, Handler: func() error {
				ctx := gui.contexts.InputPopup
				if ctx.Config != nil && ctx.Config.Locked {
					return nil
				}
				if ctx.Completion.Active && len(ctx.Completion.Items) > 0 {
					v, _ := gui.g.View(InputPopupView)
					raw := strings.TrimSpace(v.TextArea.GetUnwrappedContent())
					selected := ctx.Completion.Items[ctx.Completion.SelectedIndex]
					return gui.helpers.InputPopup().HandleEnter(raw, &selected)
				}
				return nil
			}},
			{Key: gocui.KeyCtrlS, Handler: func() error {
				ctx := gui.contexts.InputPopup
				if ctx.Config == nil || ctx.Config.OnCtrlS == nil || ctx.Config.Locked {
					return nil
				}
				v, _ := gui.g.View(InputPopupView)
				raw := strings.TrimSpace(v.TextArea.GetUnwrappedContent())
				return ctx.Config.OnCtrlS(raw)
			}},
		},
	)
}

// setupGlobalContext initializes the GlobalContext and GlobalController.
func (gui *Gui) setupGlobalContext() {
	globalCtx := context.NewGlobalContext()
	gui.contexts.Global = globalCtx
	gui.contextMgr.Register(globalCtx)

	ctrl := controllers.NewGlobalController(controllers.GlobalControllerOpts{
		Common:      gui.controllerCommon,
		GetContext:  func() *context.GlobalContext { return gui.contexts.Global },
		OnQuit:      func() error { return gui.quit(gui.g, nil) },
		OnHelp:      func() error { gui.showHelp(); return nil },
		OnPalette:   func() error { return gui.openPalette(gui.g, nil) },
		OnQuickOpen: func() error { return gui.openQuickOpen(nil, nil) },
	})
	gui.globalController = ctrl
	controllers.AttachController(ctrl)
}

// setupPaletteContext initializes the "palette" and PaletteController.
func (gui *Gui) setupPaletteContext() {
	paletteCtx := context.NewPaletteContext()
	gui.contexts.Palette = paletteCtx
	gui.contextMgr.Register(paletteCtx)

	ctrl := controllers.NewPaletteController(controllers.PaletteControllerOpts{
		GetContext:  func() *context.PaletteContext { return gui.contexts.Palette },
		OnEnter:     func() error { return gui.paletteEnter(gui.g, nil) },
		OnEsc:       func() error { return gui.paletteEsc(gui.g, nil) },
		OnListClick: func() error { return gui.paletteListClick(gui.g, nil) },
	})
	controllers.AttachController(ctrl)
}

// setupCalendarContext initializes the CalendarContext and CalendarController.
func (gui *Gui) setupCalendarContext() {
	calendarCtx := context.NewCalendarContext()
	gui.contexts.Calendar = calendarCtx
	gui.contextMgr.Register(calendarCtx)

	ctrl := controllers.NewCalendarController(controllers.CalendarControllerOpts{
		Common:     gui.controllerCommon,
		GetContext: func() *context.CalendarContext { return gui.contexts.Calendar },
	})
	controllers.AttachController(ctrl)
}

// setupPickDialogContext initializes the pick dialog context and controller.
func (gui *Gui) setupPickDialogContext() {
	pdCtx := context.NewPickDialogContext()
	gui.contexts.PickDialog = pdCtx
	gui.contextMgr.Register(pdCtx)

	ctrl := controllers.NewPickDialogController(gui.controllerCommon, func() *context.PickResultsContext {
		return gui.contexts.PickDialog
	})
	controllers.AttachController(ctrl)
}

// setupInboxBrowserContext initializes the inbox browser context and controller.
func (gui *Gui) setupInboxBrowserContext() {
	inboxCtx := context.NewInboxBrowserContext()
	gui.contexts.InboxBrowser = inboxCtx
	gui.contextMgr.Register(inboxCtx)

	ctrl := controllers.NewInboxBrowserController(
		gui.controllerCommon,
		func() *context.InboxBrowserContext { return gui.contexts.InboxBrowser },
	)
	controllers.AttachController(ctrl)
}

// setupContribContext initializes the ContribContext and ContribController.
func (gui *Gui) setupContribContext() {
	contribCtx := context.NewContribContext()
	gui.contexts.Contrib = contribCtx
	gui.contextMgr.Register(contribCtx)

	ctrl := controllers.NewContribController(controllers.ContribControllerOpts{
		Common:     gui.controllerCommon,
		GetContext: func() *context.ContribContext { return gui.contexts.Contrib },
	})
	controllers.AttachController(ctrl)
}
