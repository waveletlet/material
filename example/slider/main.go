// using example/boxes as reference to figure out how to use package

package main

import (
	"fmt"
	"log"
	"time"

	"github.com/dskinner/material"
	"github.com/dskinner/snd"
	"golang.org/x/mobile/app"
	"golang.org/x/mobile/event/lifecycle"
	"golang.org/x/mobile/event/paint"
	"golang.org/x/mobile/event/size"
	"golang.org/x/mobile/event/touch"
	"golang.org/x/mobile/gl"
)

var (
	env                    = new(material.Environment)
	slider, btnMin, btnMax *material.Button
	indicator, readout     *material.Material
	sig                    snd.Discrete
	quits                  []chan struct{}
)

func onStart(ctx gl.Context) {
	env.SetPalette(material.Palette{
		Primary: material.BlueGrey500,
		Dark:    material.BlueGrey700,
		Light:   material.BlueGrey100,
		Accent:  material.DeepOrangeA200,
	})

	quits = []chan struct{}{}

	sig = make(snd.Discrete, len(material.ExpSig))
	copy(sig, material.ExpSig)
	rsig := make(snd.Discrete, len(material.ExpSig))
	copy(rsig, material.ExpSig)
	rsig.UnitInverse()
	sig = append(sig, rsig...)
	sig.NormalizeRange(0, 1)

	ctx.Enable(gl.BLEND)
	ctx.BlendFunc(gl.SRC_ALPHA, gl.ONE_MINUS_SRC_ALPHA)

	ctx.Enable(gl.CULL_FACE)
	ctx.CullFace(gl.BACK)

	env.Load(ctx)
	env.LoadGlyphs(ctx)

	slider = env.NewButton(ctx)
	slider.SetColor(env.Palette().Dark)
	slider.OnTouch = setSlider

	btnMin = env.NewButton(ctx)
	btnMin.SetColor(env.Palette().Primary)
	btnMin.OnTouch = sliderMin
	btnMin.SetText("Min")

	btnMax = env.NewButton(ctx)
	btnMax.SetColor(env.Palette().Primary)
	btnMax.OnTouch = sliderMax
	btnMax.SetText("Max")

	indicator = env.NewMaterial(ctx)
	indicator.SetColor(env.Palette().Accent)
	indicator.Roundness = 5

	readout = env.NewMaterial(ctx)
	readout.SetColor(env.Palette().Light)
	readout.Roundness = 5

}

func setSlider(ev touch.Event) {
	if ev.Type == touch.TypeBegin {
		m := indicator.World()
		quits = append(quits, material.Animation{
			Sig:  sig,
			Dur:  1000 * time.Millisecond,
			Loop: false,
			Start: func() {
				readout.SetText(fmt.Sprintf("%v", ev.Y))
			},
			Interp: func(dt float32) {
				m[1][3] -= (m[1][3] - ev.Y) * dt
			},
			End: func() {
				m[1][3] = ev.Y
			},
		}.Do())
	}
}

func sliderMin(ev touch.Event) {
	if ev.Type == touch.TypeBegin {
		m := indicator.World()
		m2 := slider.World()
		y2 := m2[1][3]
		quits = append(quits, material.Animation{
			Sig:  sig,
			Dur:  500 * time.Millisecond,
			Loop: false,
			Start: func() {
				readout.SetText(fmt.Sprintf("%v", y2))
			},
			Interp: func(dt float32) {
				m[1][3] -= (m[1][3] - y2) * dt
				// don't use 'y' in place of m[1][3] here so the value of m[1][3] is
				// evaluated each step in the loop
			},
			End: func() {
				m[1][3] = y2
			},
		}.Do())
	}
}

func sliderMax(ev touch.Event) {
	if ev.Type == touch.TypeBegin {
		m := indicator.World()
		h1 := m[1][1]
		m2 := slider.World()
		y := m2[1][3]
		h := m2[1][1]
		end := y + h - h1
		quits = append(quits, material.Animation{
			Sig:  sig,
			Dur:  500 * time.Millisecond,
			Loop: false,
			Start: func() {
				readout.SetText(fmt.Sprintf("%v", end))
			},
			Interp: func(dt float32) {
				m[1][3] -= (m[1][3] - (end)) * dt
			},
			End: func() {
				m[1][3] = end
			},
		}.Do())
	}
}

func dumpWorld(ob *material.Material) {
	for i, sl := range ob.World() {
		log.Printf("i: %v, sl: %v\n", i, sl)
	}
}

//		m := slider.World()
//		w, x := m[0][0], m[0][3]
//    h, y := m[1][1], m[1][3]
//		z	   := m[2][3]
//		example matrix
//		[60.00,		0.0,		0.0,	1300.0]
//		[	0.00,	100.0,		0.0,	2000.0]
//		[	0.00,		0.0,		1.0,		2.00]
//		[	0.00,		0.0,		0.0,		1.00]

func onStop(ctx gl.Context) {
	env.Unload(ctx)
	for _, q := range quits {
		q <- struct{}{}
	}
}

func onLayout(sz size.Event) {
	env.SetOrtho(sz)
	env.StartLayout()

	for _, q := range quits {
		q <- struct{}{}
	}
	quits = quits[:0]

	b, p := env.Box, env.Grid.Gutter
	env.AddConstraints(
		// would be better to make element w and h variables that are related to
		// each other to keep proportions uniform on different device sizes
		slider.Width(80), slider.Height(float32(sz.HeightPx)*0.6), slider.Z(1), slider.CenterHorizontalIn(b), slider.CenterVerticalIn(b),
		indicator.Width(60), indicator.Height(100), indicator.Z(2), indicator.CenterHorizontalIn(b), indicator.CenterVerticalIn(b),
		readout.Width(600), readout.Height(200), readout.Z(2), readout.CenterHorizontalIn(b), readout.Above(btnMin.Box, p),
		btnMax.Width(600), btnMax.Height(200), btnMax.Z(3), btnMax.CenterHorizontalIn(b), btnMax.TopIn(b, p),
		btnMin.Width(600), btnMin.Height(200), btnMin.Z(3), btnMin.CenterHorizontalIn(b), btnMin.BottomIn(b, p),
	)

	//is there a reason this was in onLayout and not onStart?
	//doesn't seem to affect text size predictably either way
	//slider.SetTextHeight(material.Dp(24).Px())
	log.Println("starting layout")
	t := time.Now()
	env.FinishLayout()
	log.Printf("finished layout in %s\n", time.Now().Sub(t))

	readout.SetText(fmt.Sprintf("%v", indicator.World()[1][3]))
}

var lastpaint time.Time
var fps int

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
				//log.Printf("Touch event: %v\n", ev)
				env.Touch(ev)
			}
		}
	})
}
