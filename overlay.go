package main

/*
#cgo LDFLAGS: -lX11 -lXext -lXrender
#include <X11/Xlib.h>
#include <X11/Xutil.h>
#include <X11/Xatom.h>
#include <X11/extensions/shape.h>
#include <X11/extensions/Xrender.h>
#include <stdlib.h>
*/
import "C"
import (
	"fmt"
	"unsafe"
)

// ============ CONFIGURATION ============
const (
	TintAlpha    = 8  // Tint opacity: 0=none, 8=subtle, 25=visible, 40=strong, max 255
	BorderWidth  = 8  // Border thickness in pixels
	CornerRadius = 16 // Corner rounding in pixels
)

type Color struct {
	R, G, B float64
}

var (
	Blue    = Color{R: 0.3, G: 0.6, B: 1.0}
	Green   = Color{R: 0.45, G: 0.91, B: 0.74}
	Magenta = Color{R: 1.0, G: 0.3, B: 1.0}
	Red     = Color{R: 1.0, G: 0.2, B: 0.2}
	Yellow  = Color{R: 1.0, G: 1.0, B: 0.3}
	Cyan    = Color{R: 0.3, G: 1.0, B: 1.0}
	Purple  = Color{R: 0.6, G: 0.3, B: 1.0}
)

var (
	display *C.Display
	window  C.Window
	gc      C.GC
	screen  C.int
)

func InitOverlay() error {
	display = C.XOpenDisplay(nil)
	if display == nil {
		return fmt.Errorf("cannot open X display")
	}

	screen = C.XDefaultScreen(display)
	rootWindow := C.XRootWindow(display, screen)

	screenWidth := C.XDisplayWidth(display, screen)
	screenHeight := C.XDisplayHeight(display, screen)

	var visual *C.Visual
	var depth C.int
	var vinfo C.XVisualInfo
	var mask C.long = C.VisualScreenMask | C.VisualDepthMask | C.VisualClassMask

	vinfo.screen = screen
	vinfo.depth = 32
	vinfo.class = C.TrueColor

	var nitems C.int
	visualList := C.XGetVisualInfo(display, mask, &vinfo, &nitems)
	if visualList != nil && nitems > 0 {
		visual = visualList.visual
		depth = 32
		C.XFree(unsafe.Pointer(visualList))
	} else {
		visual = C.XDefaultVisual(display, screen)
		depth = C.XDefaultDepth(display, screen)
	}

	colormap := C.XCreateColormap(display, rootWindow, visual, C.AllocNone)

	var attrs C.XSetWindowAttributes
	attrs.override_redirect = C.True
	attrs.background_pixel = 0
	attrs.border_pixel = 0
	attrs.colormap = colormap

	window = C.XCreateWindow(
		display,
		rootWindow,
		0, 0,
		C.uint(screenWidth), C.uint(screenHeight),
		0,
		depth,
		C.uint(C.InputOutput),
		visual,
		C.CWOverrideRedirect|C.CWBackPixel|C.CWBorderPixel|C.CWColormap,
		&attrs,
	)

	C.XSelectInput(display, window, C.ExposureMask)

	var hints C.XClassHint
	className := C.CString("vimsane-overlay")
	defer C.free(unsafe.Pointer(className))
	hints.res_name = className
	hints.res_class = className
	C.XSetClassHint(display, window, &hints)

	atomName := C.CString("_NET_WM_WINDOW_TYPE")
	atomValue := C.CString("_NET_WM_WINDOW_TYPE_DOCK")
	defer C.free(unsafe.Pointer(atomName))
	defer C.free(unsafe.Pointer(atomValue))

	wmType := C.XInternAtom(display, atomName, C.False)
	wmTypeDock := C.XInternAtom(display, atomValue, C.False)
	C.XChangeProperty(display, window, wmType, C.XA_ATOM, 32, C.PropModeReplace,
		(*C.uchar)(unsafe.Pointer(&wmTypeDock)), 1)

	stateAtom := C.CString("_NET_WM_STATE")
	defer C.free(unsafe.Pointer(stateAtom))
	wmState := C.XInternAtom(display, stateAtom, C.False)

	aboveAtom := C.CString("_NET_WM_STATE_ABOVE")
	defer C.free(unsafe.Pointer(aboveAtom))
	wmStateAbove := C.XInternAtom(display, aboveAtom, C.False)

	C.XChangeProperty(display, window, wmState, C.XA_ATOM, 32, C.PropModeReplace,
		(*C.uchar)(unsafe.Pointer(&wmStateAbove)), 1)

	inputRegionAtom := C.CString("_NET_WM_WINDOW_TYPE_DOCK")
	defer C.free(unsafe.Pointer(inputRegionAtom))

	inputRegion := C.XCreateRegion()
	C.XShapeCombineRegion(display, window, C.ShapeInput, 0, 0, inputRegion, C.ShapeSet)
	C.XDestroyRegion(inputRegion)

	C.XMapWindow(display, window)
	C.XFlush(display)

	gc = C.XCreateGC(display, C.Drawable(window), 0, nil)

	return nil
}

func DrawBorder(color Color, width int) {
	if display == nil {
		return
	}

	screenWidth := C.XDisplayWidth(display, screen)
	screenHeight := C.XDisplayHeight(display, screen)
	w := C.int(width)

	C.XClearWindow(display, window)

	tintPixel := C.ulong((uint32(TintAlpha) << 24) |
		(uint32(color.R*255*0.03) << 16) |
		(uint32(color.G*255*0.03) << 8) |
		uint32(color.B*255*0.03))
	C.XSetForeground(display, gc, tintPixel)
	C.XFillRectangle(display, C.Drawable(window), gc, 0, 0, C.uint(screenWidth), C.uint(screenHeight))

	borderAlpha := uint32(255)
	borderPixel := C.ulong((borderAlpha << 24) |
		(uint32(color.R*255) << 16) |
		(uint32(color.G*255) << 8) |
		uint32(color.B*255))
	C.XSetForeground(display, gc, borderPixel)

	C.XFillRectangle(display, C.Drawable(window), gc, 0, 0, C.uint(screenWidth), C.uint(w))
	C.XFillRectangle(display, C.Drawable(window), gc, 0, 0, C.uint(w), C.uint(screenHeight))
	C.XFillRectangle(display, C.Drawable(window), gc, screenWidth-w, 0, C.uint(w), C.uint(screenHeight))
	C.XFillRectangle(display, C.Drawable(window), gc, 0, screenHeight-w, C.uint(screenWidth), C.uint(w))

	C.XFlush(display)
}

func HideWindow() {
	if display == nil {
		return
	}
	C.XUnmapWindow(display, window)
	C.XFlush(display)
}

func ShowWindow() {
	if display == nil {
		return
	}
	C.XMapWindow(display, window)
	C.XFlush(display)
}

func Cleanup() {
	if display != nil {
		if gc != nil {
			C.XFreeGC(display, gc)
		}
		if window != 0 {
			C.XDestroyWindow(display, window)
		}
		C.XCloseDisplay(display)
	}
}
