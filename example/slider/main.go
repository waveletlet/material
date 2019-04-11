// using example/boxes as reference to figure out how to use package

package main

import (
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
	env          = new(material.Environment)
	slider, btn2 *material.Button
	indicator    *material.Material
	//	boxes [9]*material.Material
	sig   snd.Discrete
	quits []chan struct{}
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
	slider.SetColor(material.Purple500)
	slider.OnTouch = setSlider

	btn2 = env.NewButton(ctx)
	btn2.SetColor(material.Teal500)
	btn2.OnTouch = sliderMin
	btn2.SetText("Reset")

	indicator = env.NewMaterial(ctx)
	indicator.SetColor(material.LightBlue500)
	indicator.Roundness = 5

}

// placeholder for function that will move the slider indicator up and down the
// slider
func setSlider(ev touch.Event) {
	if ev.Type == touch.TypeBegin {
		m := indicator.World()
		quits = append(quits, material.Animation{
			Sig:  sig,
			Dur:  1000 * time.Millisecond,
			Loop: false,
			Interp: func(dt float32) {
				m[1][3] = ev.Y * (1 - dt)
			},
			End: func() {
				m[1][3] = ev.Y
				log.Printf("ev: %v, y: %v, h: %v\n", ev, m[1][3], m[1][1])
			},
		}.Do())
	}
}

func sliderMin(ev touch.Event) {
	if ev.Type == touch.TypeBegin {
		m := indicator.World()
		y := m[1][3]
		m2 := slider.World()
		h2 := m2[1][1]
		y2 := m2[1][3]
		quits = append(quits, material.Animation{
			Sig:  sig,
			Dur:  500 * time.Millisecond,
			Loop: false,
			Start: func() {
				m[1][3] = y2
			},
			Interp: func(dt float32) {
				m[1][3] = y - y*0.6*dt
			},
			End: func() {
				//m[1][3] = y2 + h2 // top of the slider
				m[1][3] = y2
				log.Printf("ev: %v, y: %v, h2: %v, y2: %v\n", ev, y, h2, y2)
				//dumpWorld(indicator)
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
		slider.Width(50), slider.Height(float32(sz.HeightPx)*0.6), slider.Z(1), slider.CenterVerticalIn(b), slider.EndIn(b, p),
		indicator.Width(60), indicator.Height(100), indicator.Z(2), indicator.CenterVerticalIn(b), indicator.EndIn(b, p),
		btn2.Width(600), btn2.Height(200), btn2.Z(3), btn2.CenterHorizontalIn(b),
	)

	//is there a reason this was in onLayout and not onStart?
	//doesn't seem to affect text size predictably either way
	//slider.SetTextHeight(material.Dp(24).Px())
	log.Println("starting layout")
	t := time.Now()
	env.FinishLayout()
	log.Printf("finished layout in %s\n", time.Now().Sub(t))
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
