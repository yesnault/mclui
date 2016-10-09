package main

import (
	"fmt"
	"log"
	"sync"
	"time"

	marathon "github.com/gambol99/go-marathon"
	"github.com/gizak/termui"
)

var mutex = &sync.Mutex{}
var uiHeightTop = 1
var nbPanes = 3

type mclui struct {
	header      *termui.Par
	lastRefresh *termui.Par

	selectedPane             int
	selectedPaneApplications int
	current                  int

	uilists map[int]map[int]*uilist
}

type uilist struct {
	list           *termui.List
	applications   []marathon.Application
	position, page int
}

const (
	uiApplications int = iota
	//uiHome
)

func (ui *mclui) render() {
	termui.Render(termui.Body)
}

func (ui *mclui) draw(i int) {
	termui.Body.Align()
	termui.Render(termui.Body)
}

func (ui *mclui) init() {
	if err := termui.Init(); err != nil {
		panic(err)
	}

	ui.uilists = make(map[int]map[int]*uilist)

	ui.initHeader()
	ui.initLastRefresh()
	ui.showApplications()
	ui.initHandles()
}

func (ui *mclui) initHeader() {
	p := termui.NewPar("MCLUI - Marathon CLI UI | (tab) switch | (ctrl+q)uit")
	p.Height = uiHeightTop
	p.TextFgColor = termui.ColorWhite
	p.BorderTop, p.BorderLeft, p.BorderRight, p.BorderBottom = false, false, false, false
	ui.header = p
}

func (ui *mclui) initLastRefresh() {
	p := termui.NewPar("")
	p.Height = uiHeightTop
	p.TextFgColor = termui.ColorWhite
	p.BorderTop, p.BorderLeft, p.BorderRight, p.BorderBottom = false, false, false, false
	ui.lastRefresh = p
}

func (ui *mclui) prepareTopMenu() {
	termui.Body.AddRows(
		termui.NewRow(
			termui.NewCol(6, 0, ui.header),
			termui.NewCol(6, 0, ui.lastRefresh),
		),
	)
}

func (ui *mclui) showApplications() {
	ui.current = 0
	ui.selectedPane = 0
	termui.Body.Rows = nil

	ui.selectedPaneApplications = 0

	ui.initApplications()

	go func() {
		for {
			mutex.Lock()
			ui.updateApplications()
			mutex.Unlock()
			time.Sleep(10 * time.Second)
		}
	}()

	ui.colorizedPanes()

	ui.prepareTopMenu()
	termui.Body.AddRows(
		termui.NewRow(
			termui.NewCol(12, 0, ui.uilists[uiApplications][0].list),
		),
		termui.NewRow(
			termui.NewCol(12, 0, ui.uilists[uiApplications][1].list),
		),
		termui.NewRow(
			termui.NewCol(12, 0, ui.uilists[uiApplications][2].list),
		),
	)

	termui.Clear()
	ui.render()
}

func (ui *mclui) initApplications() {
	for k := 0; k < nbPanes; k++ {

		strs := []string{"[Loading...](fg-black,bg-white)"}
		ls := termui.NewList()
		ls.BorderTop, ls.BorderLeft, ls.BorderRight, ls.BorderBottom = true, false, false, false
		ls.Items = strs
		ls.ItemFgColor = termui.ColorWhite
		switch k {
		case 0:
			ls.BorderLabel = "Apps Staged"
		case 1:
			ls.BorderLabel = "Apps Running / Healthy"
		case 2:
			ls.BorderLabel = "Apps Declined / Unhealthy"
		}

		ls.Width = 25
		ls.Y = 0
		ls.Height = (termui.TermHeight() - uiHeightTop) / nbPanes

		if _, ok := ui.uilists[uiApplications]; !ok {
			ui.uilists[uiApplications] = make(map[int]*uilist)
		}
		ui.uilists[uiApplications][k] = &uilist{list: ls, position: 0, page: 0}
	}
	ui.render()
}

func (ui *mclui) initHandles() {
	// Setup handlers
	termui.Handle("/timer/1s", func(e termui.Event) {
		t := e.Data.(termui.EvtTimer)
		ui.draw(int(t.Count))
	})

	termui.Handle("/sys/kbd/C-q", func(termui.Event) {
		termui.StopLoop()
	})

	termui.Handle("/sys/kbd/<tab>", func(e termui.Event) {
		ui.switchPane()
	})

	termui.Handle("/sys/kbd/<up>", func(e termui.Event) {
		ui.move("up")
	})

	termui.Handle("/sys/kbd/<down>", func(e termui.Event) {
		ui.move("down")
	})

}

func (ui *mclui) updateApplications() {
	apps := []marathon.Application{}
	for _, c := range clients {
		applications, err := c.marathon.Applications(nil)
		if err != nil {
			log.Fatalf("Failed to list applications %s", err)
		}
		apps = append(apps, applications.Apps...)
	}

	start := time.Now().UnixNano()
	for pane := 0; pane < nbPanes; pane++ {
		ui.updateApplicationsPane(apps, pane)
		delta := int64((time.Now().UnixNano() - start) / 1000000)
		ui.lastRefresh.Text = fmt.Sprintf("%s (%dms)", time.Now().Format(time.Stamp), delta)
		ui.render()
	}
}

func (ui *mclui) updateApplicationsPane(apps []marathon.Application, pane int) {
	var strs []string
	for _, app := range apps {
		toAdd := false
		if pane == 0 && app.TasksStaged > 0 { //staged
			toAdd = true
		} else if pane == 1 && (app.TasksHealthy > 0 || *app.Instances >= app.TasksRunning) { //healthy / running
			toAdd = true
		} else if pane == 3 && (app.TasksUnhealthy > 0 || *app.Instances < app.TasksRunning) { // unhealthy / declined
			toAdd = true
		}
		if toAdd {
			strs = append(strs, ui.formatApplication(app, true))
			ui.uilists[uiApplications][pane].applications = append(ui.uilists[uiApplications][pane].applications, app)
		}
	}

	nbPerPage := ui.getNbPerPage()
	page := ui.uilists[uiApplications][pane].page
	skip := page * nbPerPage
	if len(strs) > 0 && skip < len(ui.uilists[uiApplications][pane].applications) {
		limit := (page + 1) * nbPerPage
		if limit > len(strs) {
			limit = len(strs)
		}
		ui.uilists[uiApplications][pane].list.Items = strs[skip:limit]
	} else {
		ui.uilists[uiApplications][pane].list.Items = []string{}
	}

	if ui.selectedPaneApplications == pane {
		ui.addMarker(pane)
	}
	ui.render()
}

func (ui *mclui) move(direction string) {

	ui.colorizedPanes()
	if direction == "up" {
		ui.moveUP(ui.uilists[ui.selectedPane][ui.selectedPaneApplications])
	} else {
		ui.moveDown(ui.uilists[ui.selectedPane][ui.selectedPaneApplications])
	}
}

func (ui *mclui) moveUP(uil *uilist) {
	p := uil.position
	uil.position--
	if uil.position < 0 {
		uil.page--
		if uil.page < 0 {
			uil.page = 0
			uil.position = 0
		} else {
			ui.updateApplicationsPane(uil.applications, ui.selectedPaneApplications)
			uil.position = len(uil.list.Items) - 1
		}
	}
	if p != uil.position {
		ui.removeMarker(uil, ui.selectedPaneApplications, p)
		ui.addMarker(ui.selectedPaneApplications)
	}
}

func (ui *mclui) moveDown(uil *uilist) {
	p := uil.position
	uil.position++

	skip := uil.page*ui.getNbPerPage() + uil.position
	if len(uil.applications) <= skip {
		return
	}

	if uil.position >= len(uil.list.Items) && ui.getNbPerPage() == len(uil.list.Items) {
		uil.page++
		ui.updateApplicationsPane(uil.applications, ui.selectedPaneApplications)
		uil.position = 0
	} else if ui.getNbPerPage() > len(uil.list.Items) && uil.position >= len(uil.list.Items) {
		uil.position--
	}
	if p != uil.position {
		ui.removeMarker(uil, ui.selectedPaneApplications, p)
		ui.addMarker(ui.selectedPaneApplications)
	}
}

func (ui *mclui) addMarker(selectedPane int) {
	uil := ui.uilists[uiApplications][ui.selectedPaneApplications]
	if uil == nil || uil.list.Items == nil || uil.position < 0 || uil.position >= len(uil.list.Items) {
		return
	}

	if len(uil.applications) > uil.position {
		skip := uil.page * ui.getNbPerPage()
		app := uil.applications[skip+uil.position]
		uil.list.Items[uil.position] = fmt.Sprintf("[%s](bg-green)", ui.formatApplication(app, false))
	}
	ui.render()
}

func (ui *mclui) removeMarker(uil *uilist, selectedPane, pos int) {
	if pos < 0 || pos >= len(uil.list.Items) {
		return
	}
	if pos >= len(uil.applications) {
		return
	}
	skip := uil.page * ui.getNbPerPage()
	app := uil.applications[skip+pos]
	uil.list.Items[pos] = ui.formatApplication(app, true)
	ui.render()
}

func (ui *mclui) formatApplication(app marathon.Application, withColor bool) string {

	return fmt.Sprintf("%s Staged:%d Running:%d Healthy:%d Unhealthy:%d v:%s changed:%s",
		app.ID,
		app.TasksStaged,
		app.TasksRunning,
		app.TasksHealthy,
		app.TasksUnhealthy,
		app.Version,
		app.VersionInfo.LastConfigChangeAt,
	)
}

func (ui *mclui) getNbPerPage() int {
	return (termui.TermHeight() - uiHeightTop - nbPanes) / nbPanes
}

func (ui *mclui) switchPane() {
	ui.removeMarker(ui.uilists[uiApplications][ui.selectedPaneApplications], ui.selectedPaneApplications, ui.uilists[uiApplications][ui.selectedPaneApplications].position)
	ui.selectedPaneApplications++
	if ui.selectedPaneApplications >= nbPanes {
		ui.selectedPaneApplications = 0
	}
	ui.uilists[uiApplications][ui.selectedPaneApplications].position = 0
	ui.addMarker(ui.selectedPaneApplications)
	ui.colorizedPanes()
}

func (ui *mclui) colorizedPanes() {
	for k := 0; k < nbPanes; k++ {
		ui.uilists[uiApplications][k].list.BorderFg = termui.ColorWhite
	}
	ui.uilists[ui.selectedPane][ui.selectedPaneApplications].list.BorderFg = termui.ColorRed
}

func runUI() {
	ui := &mclui{}
	ui.init()
	ui.draw(1)

	defer termui.Close()
	termui.Loop()
}
