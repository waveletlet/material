package main

import (
	"log"
	"time"

	"github.com/dskinner/material"
	"github.com/dskinner/material/icon"
	"golang.org/x/mobile/app"
	"golang.org/x/mobile/event/lifecycle"
	"golang.org/x/mobile/event/paint"
	"golang.org/x/mobile/event/size"
	"golang.org/x/mobile/event/touch"
	"golang.org/x/mobile/gl"
)

var (
	env                      = new(material.Environment)
	showBtn, hideBtn, b1, b2 *material.Button
	mnu                      *material.Menu
)

func onStart(ctx gl.Context) {
	env.SetPalette(material.Palette{
		Primary: material.BlueGrey500,
		Dark:    material.BlueGrey700,
		Light:   material.BlueGrey100,
		Accent:  material.DeepOrangeA200,
	})

	ctx.Enable(gl.BLEND)
	ctx.BlendFunc(gl.SRC_ALPHA, gl.ONE_MINUS_SRC_ALPHA)

	ctx.Enable(gl.CULL_FACE)
	ctx.CullFace(gl.BACK)

	env.Load(ctx)
	env.LoadGlyphs(ctx)

	showBtn = env.NewButton(ctx)
	showBtn.SetTextColor(material.White)
	showBtn.SetIcon(icon.NavigationMenu)
	showBtn.SetText("Show Menu")
	showBtn.OnTouch = showMenu

	hideBtn = env.NewButton(ctx)
	hideBtn.SetTextColor(material.White)
	hideBtn.SetIcon(icon.NavigationMenu)
	hideBtn.SetText("Hide Menu")
	hideBtn.OnTouch = hideMenu

	b1 = env.NewButton(ctx)
	b1.SetTextColor(material.White)
	b1.SetText("Menu thing A")

	b2 = env.NewButton(ctx)
	b2.SetTextColor(material.White)
	b2.SetText("Menu thing B")

	mnu = env.NewMenu(ctx)
	mnu.AddAction(b1)
	mnu.AddAction(b2)
}

// Showing/hiding menu doesn't seem to work out of the box
// Flickery animation shows and shadowy area that's probably supposed to be the
// menu button disappears. Is the functionality done or am I placing it wrong
// in the env.AddConstraints?
// 'mnu.ShowAt(&env.View)' completely disappears the menu
// Is env.View the right argument? It's looking for a *f32.Mat4.
func showMenu(ev touch.Event) {
	log.Printf("Showing Menu! %v\n", ev)
	mnu.Show()
	//mnu.ShowAt(&env.View)
}

func hideMenu(ev touch.Event) {
	log.Printf("Hiding menu! %v\n", ev)
	mnu.Hide()
}

func onStop(ctx gl.Context) {
	env.Unload(ctx)
}

func onLayout(sz size.Event) {
	env.SetOrtho(sz)
	env.StartLayout()

	env.AddConstraints(
		showBtn.Width(900), showBtn.Height(200), showBtn.Z(1), showBtn.CenterVerticalIn(env.Box), showBtn.CenterHorizontalIn(env.Box),
		hideBtn.Width(900), hideBtn.Height(200), hideBtn.Z(1), hideBtn.Below(env.Box, env.Grid.Gutter), hideBtn.CenterHorizontalIn(env.Box),
	)

	for _, cns := range mnu.Constraints(env) {
		env.AddConstraints(cns)
	}

	log.Println("starting layout")
	t := time.Now()
	env.FinishLayout()
	log.Printf("finished layout in %s\n", time.Now().Sub(t))
}

var fps int
var lastpaint time.Time

func onPaint(ctx gl.Context) {
	ctx.ClearColor(material.BlueGrey100.RGBA())
	ctx.Clear(gl.COLOR_BUFFER_BIT)
	env.Draw(ctx)
	now := time.Now()
	fps = int(time.Second / now.Sub(lastpaint))
	lastpaint = now
}

func main() {
	app.Main(func(a app.App) {
		var glctx gl.Context
		for ev := range a.Events() {
			switch ev := a.Filter(ev).(type) {
			case lifecycle.Event:
				switch ev.Crosses(lifecycle.StageVisible) {
				case lifecycle.CrossOn:
					glctx = ev.DrawContext.(gl.Context)
					onStart(glctx)
					a.Send(paint.Event{})
				case lifecycle.CrossOff:
					onStop(glctx)
					glctx = nil
				}
			case size.Event:
				if glctx == nil {
					a.Send(ev) // republish event until onStart is called
				} else {
					onLayout(ev)
				}
			case paint.Event:
				if glctx == nil || ev.External {
					continue
				}
				onPaint(glctx)
				a.Publish()
				a.Send(paint.Event{})
			case touch.Event:
				env.Touch(ev)
			}
		}
	})
}
