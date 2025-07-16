package main

import (
	"fmt"
	"image/color"
	"log"
	"os"
	"strings"

	"gioui.org/layout"
	"gioui.org/unit"
	"gioui.org/widget"
	"gioui.org/x/component"

	// "gioui.org/widget"
	// "golang.org/x/exp/shiny/materialdesign/icons"
	// "golang.org/x/exp/shiny/unit"

	"gioui.org/app"
	"gioui.org/font/gofont"
	"gioui.org/op"
	"gioui.org/text"
	"gioui.org/widget/material"
)

type (
	C = layout.Context
	D = layout.Dimensions
)
type IconItem struct {
	name  string
	data  []byte
	click widget.Clickable
}

func (p *IconItem) Layout(gtx C, th *material.Theme) D {
	icon, _ := widget.NewIcon(p.data)
	if p.click.Clicked(gtx) {
		fmt.Printf("%v clicked\n", p.name)
	}
	return material.IconButton(th, &p.click, icon, p.name).Layout(gtx)
}

type UI struct {
	inputAlignment         text.Alignment
	textField, resultField component.TextField
	// click                  widget.Clickable
	search_txt            string // 用于存储搜索文本
	iconitems []IconItem //用于显示的icon
	iconitems_all []IconItem  // 所有的icon
}

func main() {
	ui := new(UI)
	go func() {
		w := new(app.Window)
		w.Option(
			// app.Size(unit.Dp(800), unit.Dp(400)),
			app.Title("Gio Icon Viewer"),
		)
		if err := ui.loop(w); err != nil {
			log.Fatal(err)
		}
		os.Exit(0)
	}()
	app.Main()
}

func (ui *UI) loop(w *app.Window) error {
	th := material.NewTheme()
	th.Shaper = text.NewShaper(text.WithCollection(gofont.Collection()))
	var ops op.Ops

	ui.iconitems_all = InitIconItems()
	ui.iconitems = ui.iconitems_all // 默认显示所有icon
	fmt.Printf("len(items): %d\n", len(ui.iconitems_all))

	for {
		switch e := w.Event().(type) {
		case app.DestroyEvent:
			return e.Err
		case app.FrameEvent:
			gtx := app.NewContext(&ops, e)
			gtx.Constraints.Min = gtx.Constraints.Max // 使布局充满窗口

			//每行个数
			n := int(gtx.Constraints.Max.X / 50)
			// fmt.Println("显示列数n:", n)

			layout.Flex{
				Axis:      layout.Vertical,
				Spacing:   layout.SpaceEvenly,
				Alignment: layout.Start,
			}.Layout(gtx,
				layout.Rigid(func(gtx layout.Context) layout.Dimensions {
					return layout.Flex{Axis: layout.Horizontal, Alignment: layout.Middle}.Layout(gtx,
						layout.Flexed(0.3, func(gtx layout.Context) layout.Dimensions {
							ui.textField.Alignment = ui.inputAlignment
							ui.textField.SingleLine = true
							ui.textField.Submit = true
							ui.textField.Prefix = func(gtx C) D {
								th := *th
								th.Palette.Fg = color.NRGBA{R: 100, G: 100, B: 100, A: 255}
								return material.Label(&th, th.TextSize, "※").Layout(gtx)
							}
							// 监听输入框的提交事件
							txt := ui.textField.Text()
							if ui.search_txt != txt { // 如果文本变化了，重新过滤
								ui.search_txt = txt // 更新搜索文本
								fmt.Printf("Search submitted: %s\n", txt)
								// ui.textField.Editor.Submit = false // 重置提交状态
								if txt == "" {
									ui.iconitems = ui.iconitems_all
								} else {
									ui.filterIconWidget(txt)
								}
							}
							
							return ui.textField.Layout(gtx, th, "Search")
						}),
						// layout.Rigid(func(gtx layout.Context) layout.Dimensions {
						// 	if ui.click.Clicked(gtx) {
						// 		txt := ui.textField.Text()
						// 		fmt.Printf("Search clicked: %s\n", txt)
						// 		if txt == "" {
						// 			ui.iconitems = ui.iconitems_all
						// 		} else {
						// 			ui.filterIconWidget(txt)
						// 		}
						// 	}
						// 	return material.Button(th, &ui.click, "Search").Layout(gtx)
						// }),
						layout.Flexed(0.3, func(gtx layout.Context) layout.Dimensions {
							ui.resultField.Alignment = ui.inputAlignment
							ui.resultField.SingleLine = true
							return ui.resultField.Layout(gtx, th, "Click an icon to show its name")
						}),
					)
				}),
				layout.Flexed(1, func(gtx layout.Context) layout.Dimensions {
					return layout.UniformInset(unit.Dp(10)).Layout(gtx, func(gtx layout.Context) layout.Dimensions {
						return layout.Flex{Axis: layout.Vertical}.Layout(gtx,
							ui.layoutAllIconWidget(th, gtx, n)...,
						)
					})
				}),
			)

			e.Frame(gtx.Ops)
		}
	}
}

func (ui *UI) layoutAllIconWidget(th *material.Theme, gtx layout.Context, columns int) []layout.FlexChild {
	var items []layout.FlexChild

	total := len(ui.iconitems)
	rows := total / columns         // 共多少行（不包括最后一行）
	lastRowItems := total % columns // 最后一行的个数

	for i := 0; i < rows; i++ {
		var cells []layout.FlexChild
		for j := i * columns; j < (i+1)*columns; j++ {
			iconitem := &ui.iconitems[j] // 注意：不要忘记&
			if iconitem.click.Clicked(gtx) {
				fmt.Printf("%v clicked\n", iconitem.name)
				ui.resultField.SetText(iconitem.name)
			}
			cells = append(cells, layout.Rigid(func(gtx layout.Context) layout.Dimensions {
				return iconitem.Layout(gtx, th)
			}))
		}
		row := layout.Rigid(func(gtx layout.Context) layout.Dimensions {
			return layout.Flex{Axis: layout.Horizontal}.Layout(gtx,
				cells...,
			)
		})
		items = append(items, row)
	}

	// 最后一行
	if lastRowItems > 0 {
		var cells []layout.FlexChild
		for j := rows * columns; j < total; j++ {
			iconitem := &ui.iconitems[j]
			if iconitem.click.Clicked(gtx) {
				fmt.Printf("%v clicked\n", iconitem.name)
				ui.resultField.SetText(iconitem.name)
			}
			cells = append(cells, layout.Rigid(func(gtx layout.Context) layout.Dimensions {
				return iconitem.Layout(gtx, th)
			}))
		}
		row := layout.Rigid(func(gtx layout.Context) layout.Dimensions {
			return layout.Flex{Axis: layout.Horizontal}.Layout(gtx,
				cells...,
			)
		})
		items = append(items, row)
	}

	return items
}

func (ui *UI) filterIconWidget(txt string) {
	fmt.Printf("filterIconWidget: %s\n", txt)
	var iconitems []IconItem = make([]IconItem, 0)
	t := strings.ToLower(txt)
	for _, item := range ui.iconitems_all {
		if strings.Contains(strings.ToLower(item.name), t) {
			iconitems = append(iconitems, item)
		}
	}
	fmt.Printf("len(iconitems): %d\n", len(iconitems))
	ui.iconitems = iconitems
}
