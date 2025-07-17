package main

import (
	"fmt"
	"image/color"
	"log"
	"os"
	"strings"

	"gioui.org/app"
	"gioui.org/font/gofont"
	"gioui.org/layout"
	"gioui.org/op"
	"gioui.org/text"
	"gioui.org/widget"
	"gioui.org/widget/material"
	"gioui.org/x/component"
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
	search_txt    string     // 用于存储搜索文本
	iconitems     []IconItem //用于显示的icon
	iconitems_all []IconItem // 所有的icon

	list      widget.List
	scrollbar widget.Scrollbar
}

func main() {
	ui := new(UI)
	go func() {
		w := new(app.Window)
		w.Option(
			// app.Size(unit.Dp(800), unit.Dp(400)),
			app.Title("Gio Icon Viewer(by Joyeah)"),
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
	ui.list.Axis = layout.Vertical

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
			}.Layout(gtx,
				layout.Rigid(func(gtx layout.Context) layout.Dimensions {
					return layout.Flex{Axis: layout.Horizontal, Alignment: layout.Middle}.Layout(gtx,
						layout.Flexed(1, func(gtx layout.Context) layout.Dimensions {
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

						
						layout.Flexed(1, func(gtx layout.Context) layout.Dimensions {
							ui.resultField.Alignment = ui.inputAlignment
							ui.resultField.SingleLine = true
							return ui.resultField.Layout(gtx, th, "Click an icon to show its name")
						}),
					)
				}),
				layout.Flexed(1, func(gtx layout.Context) layout.Dimensions {
					// 计算每行显示的icon个数
					total := len(ui.iconitems)
					listLength := total / n   // 共多少行（不包括最后一行）
					lastRowItems := total % n // 最后一行的个数
					if lastRowItems > 0 {
						listLength += 1 // 如果有最后一行，则加1
					}

					return layout.Flex{Axis: layout.Vertical}.Layout(gtx,
						layout.Flexed(99, func(gtx C) D {

							return material.List(th, &ui.list).Layout(gtx, listLength, func(gtx C, i int) D {
								// return rows[i]
								// return material.Label(th, unit.Sp(16), fmt.Sprintf("Item %d", i)).Layout(gtx)
								var cells = ui.layoutRowWidget(th, gtx, n, i, listLength, lastRowItems)

								return layout.Flex{Axis: layout.Horizontal}.Layout(gtx,
									cells...,
								)

							})

						}),
						layout.Flexed(1, func(gtx C) D {
							start, end := float32(ui.list.Position.First), float32(ui.list.Position.First+ui.list.Position.Count)
							return material.Scrollbar(th, &ui.scrollbar).Layout(gtx, layout.Vertical, start, end)
						}),
					)

				}),
			)

			e.Frame(gtx.Ops)
		}
	}
}

func (ui *UI) layoutRowWidget(th *material.Theme, gtx C, n int, i int, listLength, lastRowItems int) []layout.FlexChild {
	var cells []layout.FlexChild
	// fmt.Printf("listLength: %d, lastRowItems: %d, i: %d\n", listLength, lastRowItems, i)
	if i < listLength-1 {
		cells = make([]layout.FlexChild, n)
		for j := 0; j < n; j++ {
			iconitem := &ui.iconitems[i*n + j]
			if iconitem.click.Clicked(gtx) {
				fmt.Printf("%v clicked\n", iconitem.name)
				ui.resultField.SetText(iconitem.name)
			}
			cells[j] = layout.Flexed(1, func(gtx C) D {
				return iconitem.Layout(gtx, th)
			})
		}
	} else {
		// 最后一行
		if i == listLength-1 && lastRowItems > 0 {
			cells = make([]layout.FlexChild, lastRowItems)
		} else {
			lastRowItems = n 
			cells = make([]layout.FlexChild, n)
		}
		for j := 0; j < lastRowItems; j++ {
			// fmt.Printf("i: %d, j: %d, n: %d\n", i, j, n)
			iconitem := &ui.iconitems[i*n + j]
			if iconitem.click.Clicked(gtx) {
				fmt.Printf("%v clicked\n", iconitem.name)
				ui.resultField.SetText(iconitem.name)
			}
			cells[j] = layout.Rigid(func(gtx C) D {
				return ui.iconitems[i*n + j].Layout(gtx, th)
			})
		}
	}
	return cells
}


func (ui *UI) filterIconWidget(txt string) {
	// fmt.Printf("filterIconWidget: %s\n", txt)
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
