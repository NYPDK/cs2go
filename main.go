package main

import (
	"fmt"
	"runtime"
	"syscall"
	"time"
	"unsafe"

	"github.com/lxn/win"
	"golang.org/x/sys/windows"
)

type Matrix [4][4]float32

type Vector3 struct {
	X float32
	Y float32
	Z float32
}

var (
	user32                     = windows.NewLazySystemDLL("user32.dll")
	gdi32                      = windows.NewLazySystemDLL("gdi32.dll")
	getSystemMetrics           = user32.NewProc("GetSystemMetrics")
	setLayeredWindowAttributes = user32.NewProc("SetLayeredWindowAttributes")
	showCursor                 = user32.NewProc("ShowCursor")
	createBrush                = gdi32.NewProc("CreateBrushIndirect")
	createCompatibleDC         = gdi32.NewProc("CreateCompatibleDC")
	createSolidBrush           = gdi32.NewProc("CreateSolidBrush")
	createPen                  = gdi32.NewProc("CreatePen")
)

func init() {
	// Ensure main() runs on the main thread.
	runtime.LockOSThread()
}

func logAndSleep(err error) {
	fmt.Println(err)
	time.Sleep(5 * time.Second)
}

func worldToScreen(viewMatrix Matrix, position Vector3) (float32, float32) {
	var screenX float32
	var screenY float32
	screenX = viewMatrix[0][0]*position.X + viewMatrix[0][1]*position.Y + viewMatrix[0][2]*position.Z + viewMatrix[0][3]
	screenY = viewMatrix[1][0]*position.X + viewMatrix[1][1]*position.Y + viewMatrix[1][2]*position.Z + viewMatrix[1][3]
	w := viewMatrix[3][0]*position.X + viewMatrix[3][1]*position.Y + viewMatrix[3][2]*position.Z + viewMatrix[3][3]
	if w < 0.01 {
		return -1, -1
	}
	invw := 1.0 / w
	screenX *= invw
	screenY *= invw
	width, _, _ := getSystemMetrics.Call(0)
	height, _, _ := getSystemMetrics.Call(1)
	widthFloat := float32(width)
	heightFloat := float32(height)
	x := widthFloat / 2
	y := heightFloat / 2
	x += 0.5*screenX*widthFloat + 0.5
	y -= 0.5*screenY*heightFloat + 0.5
	return x, y
}

func getEntitiesInfo(procHandle windows.Handle, clientDll uintptr, screenWidth uintptr, screenHeight uintptr) ([][4][2]float32, []int32, []int32) {
	var localPlayerP uintptr
	err := read(procHandle, clientDll+dwLocalPlayerPawn, &localPlayerP)
	if err != nil {
		return nil, nil, nil
	}
	var entityList uintptr
	err = read(procHandle, clientDll+dwEntityList, &entityList)
	if err != nil {
		return nil, nil, nil
	}
	var listEntry uintptr
	var entityController uintptr
	var entityControllerPawn uintptr
	var entityPawn uintptr
	var entityTeam int32
	var entityHealth int32
	var entityOrigin Vector3
	var viewMatrix Matrix

	var entRects [][4][2]float32
	var entityTeams []int32
	var entityHealths []int32
	for i := 0; i < 64; i++ {
		// listEntry
		err := read(procHandle, entityList+uintptr((8*(i&0x7FFF)>>9)+16), &listEntry)
		if err != nil {
			logAndSleep(err)
			return nil, nil, nil
		}
		if listEntry == 0 {
			continue
		}
		// entityController
		err = read(procHandle, listEntry+uintptr(120)*uintptr(i&0x1FF), &entityController)
		if err != nil {
			logAndSleep(err)
			return nil, nil, nil
		}
		if entityController == 0 {
			continue
		}
		// entityControllerPawn
		err = read(procHandle, entityController+m_hPlayerPawn, &entityControllerPawn)
		if err != nil {
			logAndSleep(err)
			return nil, nil, nil
		}
		if entityControllerPawn == 0 {
			continue
		}
		// listEntry
		err = read(procHandle, entityList+uintptr(0x8*((entityControllerPawn&0x7FFF)>>9)+16), &listEntry)
		if err != nil {
			logAndSleep(err)
			return nil, nil, nil
		}
		if listEntry == 0 {
			continue
		}
		// entityPawn
		err = read(procHandle, listEntry+uintptr(120)*uintptr(entityControllerPawn&0x1FF), &entityPawn)
		if err != nil {
			logAndSleep(err)
			return nil, nil, nil
		}
		if entityPawn == 0 {
			continue
		}
		if entityPawn == localPlayerP {
			continue
		}
		// entityTeam
		err = read(procHandle, entityPawn+m_iTeamNum, &entityTeam)
		if err != nil {
			logAndSleep(err)
			return nil, nil, nil
		}
		if entityTeam == 0 {
			continue
		}
		// entityHealth
		err = read(procHandle, entityPawn+m_iHealth, &entityHealth)
		if err != nil {
			logAndSleep(err)
			return nil, nil, nil
		}
		if entityHealth <= 0 {
			continue
		}
		// entityOrigin
		err = read(procHandle, entityPawn+m_vOldOrigin, &entityOrigin)
		if err != nil {
			logAndSleep(err)
			return nil, nil, nil
		}
		entityHead := Vector3{X: entityOrigin.X, Y: entityOrigin.Y, Z: entityOrigin.Z + 70.0}
		// viewMatrix
		err = read(procHandle, clientDll+dwViewMatrix, &viewMatrix)
		if err != nil {
			logAndSleep(err)
			return nil, nil, nil
		}
		screenPosHeadX, screenPosHeadY := worldToScreen(viewMatrix, entityHead)
		_, screenPosFeetY := worldToScreen(viewMatrix, entityOrigin)
		if screenPosHeadX <= -1 || screenPosHeadY <= -1 || screenPosHeadX >= float32(screenWidth) || screenPosFeetY >= float32(screenHeight) {
			continue
		}
		boxHeight := screenPosFeetY - screenPosHeadY

		entRects = append(entRects, [4][2]float32{
			{screenPosHeadX - boxHeight/4, screenPosHeadY},
			{screenPosHeadX + boxHeight/4, screenPosHeadY},
			{screenPosHeadX + boxHeight/4, screenPosFeetY},
			{screenPosHeadX - boxHeight/4, screenPosFeetY}})
		entityTeams = append(entityTeams, entityTeam)
		entityHealths = append(entityHealths, entityHealth)
	}
	return entRects, entityTeams, entityHealths
}

func drawRect(hdc win.HDC, tPen uintptr, gPen uintptr, oPen uintptr, rect [4][2]float32, hp int32) {
	win.SelectObject(hdc, win.HGDIOBJ(oPen))
	win.MoveToEx(hdc, int(rect[0][0])-1, int(rect[0][1])-1, nil)
	win.LineTo(hdc, int32(rect[1][0])+1, int32(rect[1][1])-1)
	win.LineTo(hdc, int32(rect[2][0])+1, int32(rect[2][1])+1)
	win.LineTo(hdc, int32(rect[3][0])-1, int32(rect[3][1])+1)
	win.LineTo(hdc, int32(rect[0][0])-1, int32(rect[0][1])-1)
	win.MoveToEx(hdc, int(rect[0][0])+1, int(rect[0][1])+1, nil)
	win.LineTo(hdc, int32(rect[1][0])-1, int32(rect[1][1])+1)
	win.LineTo(hdc, int32(rect[2][0])-1, int32(rect[2][1])-1)
	win.LineTo(hdc, int32(rect[3][0])+1, int32(rect[3][1])-1)
	win.LineTo(hdc, int32(rect[0][0])+1, int32(rect[0][1])+1)
	win.SelectObject(hdc, win.HGDIOBJ(tPen))
	win.MoveToEx(hdc, int(rect[0][0]), int(rect[0][1]), nil)
	win.LineTo(hdc, int32(rect[1][0]), int32(rect[1][1]))
	win.LineTo(hdc, int32(rect[2][0]), int32(rect[2][1]))
	win.LineTo(hdc, int32(rect[3][0]), int32(rect[3][1]))
	win.LineTo(hdc, int32(rect[0][0]), int32(rect[0][1]))
	win.SelectObject(hdc, win.HGDIOBJ(gPen))
	win.MoveToEx(hdc, int(rect[0][0])-4, int(rect[3][1])+1-int(float64(int(rect[3][1])+1-int(rect[0][1]))*float64(hp)/100.0), nil)
	win.LineTo(hdc, int32(rect[3][0])-4, int32(rect[3][1])+1)
	win.SelectObject(hdc, win.HGDIOBJ(oPen))
	win.MoveToEx(hdc, int(rect[0][0])-5, int(rect[0][1])-1, nil)
	win.LineTo(hdc, int32(rect[0][0])-5, int32(rect[3][1])+1)
	win.LineTo(hdc, int32(rect[0][0])-3, int32(rect[3][1])+1)
	win.LineTo(hdc, int32(rect[0][0])-3, int32(rect[0][1])-1)
	win.LineTo(hdc, int32(rect[0][0])-5, int32(rect[0][1])-1)
}

func windowProc(hwnd win.HWND, msg uint32, wParam, lParam uintptr) uintptr {
	switch msg {
	case win.WM_TIMER:
		return 0
	case win.WM_DESTROY:
		win.PostQuitMessage(0)
		return 0
	default:
		return win.DefWindowProc(hwnd, msg, wParam, lParam)
	}
}

func initWindow(screenWidth uintptr, screenHeight uintptr) win.HWND {

	className, err := windows.UTF16PtrFromString("cs2goWindow")
	if err != nil {
		logAndSleep(err)
		return 0
	}
	windowTitle, err := windows.UTF16PtrFromString("cs2go")
	if err != nil {
		logAndSleep(err)
		return 0
	}

	// Register window class
	wc := win.WNDCLASSEX{
		CbSize:        uint32(unsafe.Sizeof(win.WNDCLASSEX{})),
		Style:         win.CS_HREDRAW | win.CS_VREDRAW,
		LpfnWndProc:   syscall.NewCallback(windowProc),
		CbWndExtra:    0,
		HInstance:     win.GetModuleHandle(nil),
		HIcon:         win.LoadIcon(0, (*uint16)(unsafe.Pointer(uintptr(win.IDI_APPLICATION)))),
		HCursor:       win.LoadCursor(0, (*uint16)(unsafe.Pointer(uintptr(win.IDC_ARROW)))),
		HbrBackground: win.COLOR_WINDOW,
		LpszMenuName:  nil,
		LpszClassName: className,
		HIconSm:       win.LoadIcon(0, (*uint16)(unsafe.Pointer(uintptr(win.IDI_APPLICATION)))),
	}
	bgBrush, _, _ := createSolidBrush.Call(uintptr(0x000000))
	wc.HbrBackground = win.HBRUSH(bgBrush)

	if atom := win.RegisterClassEx(&wc); atom == 0 {
		fmt.Println("Error registering window class:", win.GetLastError())
		return 0
	}

	// Create window
	hInstance := win.GetModuleHandle(nil)
	hwnd := win.CreateWindowEx(
		win.WS_EX_TOPMOST|win.WS_EX_NOACTIVATE|win.WS_EX_LAYERED,
		className,
		windowTitle,
		win.WS_POPUP,
		0,
		0,
		int32(screenWidth),
		int32(screenHeight),
		0,
		0,
		hInstance,
		nil,
	)
	if hwnd == 0 {
		fmt.Println("Error creating window:", win.GetLastError())
		return 0
	}

	result, _, _ := setLayeredWindowAttributes.Call(uintptr(hwnd), 0x000000, 0, 0x00000001)
	if result == 0 {
		logAndSleep(fmt.Errorf("error setting layered window attributes: %v", win.GetLastError()))
	}

	showCursor.Call(0)

	// Show window
	win.ShowWindow(hwnd, win.SW_SHOWDEFAULT)
	return hwnd
}

func main() {
	screenWidth, _, _ := getSystemMetrics.Call(0)
	screenHeight, _, _ := getSystemMetrics.Call(1)

	hwnd := initWindow(screenWidth, screenHeight)
	if hwnd == 0 {
		logAndSleep(fmt.Errorf("error initializing window"))
		return
	}

	pid, err := findProcessId("cs2.exe")
	if err != nil {
		logAndSleep(err)
		return
	}

	clientDll, err := getModuleBaseAddress(pid, "client.dll")
	if err != nil {
		logAndSleep(err)
		return
	}

	procHandle, err := getProcessHandle(pid)
	if err != nil {
		logAndSleep(err)
		return
	}

	hdc := win.GetDC(hwnd)
	if hdc == 0 {
		logAndSleep(fmt.Errorf("error getting device context: %v", win.GetLastError()))
		return
	}
	fmt.Println(hdc)

	bgBrush, _, _ := createSolidBrush.Call(uintptr(0x000000))
	if bgBrush == 0 {
		fmt.Println("Error creating brush:", win.GetLastError())
		return
	}
	defer win.DeleteObject(win.HGDIOBJ(bgBrush))
	redPen, _, _ := createPen.Call(win.PS_SOLID, 1, 0x0000FF)
	if redPen == 0 {
		fmt.Println("Error creating pen:", win.GetLastError())
		return
	}
	defer win.DeleteObject(win.HGDIOBJ(redPen))
	greenPen, _, _ := createPen.Call(win.PS_SOLID, 1, 0x00F800)
	if greenPen == 0 {
		fmt.Println("Error creating pen:", win.GetLastError())
		return
	}
	defer win.DeleteObject(win.HGDIOBJ(greenPen))
	bluePen, _, _ := createPen.Call(win.PS_SOLID, 1, 0xFF0000)
	if bluePen == 0 {
		fmt.Println("Error creating pen:", win.GetLastError())
		return
	}
	defer win.DeleteObject(win.HGDIOBJ(bluePen))
	outlinePen, _, _ := createPen.Call(win.PS_SOLID, 1, 0x000001)
	if outlinePen == 0 {
		fmt.Println("Error creating pen:", win.GetLastError())
		return
	}
	defer win.DeleteObject(win.HGDIOBJ(outlinePen))

	win.SetTimer(hwnd, 1, 1, 0)
	var msg win.MSG
	for win.GetMessage(&msg, 0, 0, 0) > 0 {
		start := time.Now()
		memhdc, _, _ := createCompatibleDC.Call(uintptr(hdc))
		memBitmap := win.CreateCompatibleBitmap(hdc, int32(screenWidth), int32(screenHeight))
		win.SelectObject(win.HDC(memhdc), win.HGDIOBJ(memBitmap))
		win.SelectObject(win.HDC(memhdc), win.HGDIOBJ(bgBrush))
		win.TranslateMessage(&msg)
		win.DispatchMessage(&msg)

		rects, teams, healths := getEntitiesInfo(procHandle, clientDll, screenWidth, screenHeight)
		for i, rect := range rects {
			if teams[i] == 2 {
				drawRect(win.HDC(memhdc), redPen, greenPen, outlinePen, rect, healths[i])
			} else {
				drawRect(win.HDC(memhdc), bluePen, greenPen, outlinePen, rect, healths[i])
			}
		}
		win.BitBlt(hdc, 0, 0, int32(screenWidth), int32(screenHeight), win.HDC(memhdc), 0, 0, win.SRCCOPY)

		// Delete the memory bitmap and device context
		win.DeleteObject(win.HGDIOBJ(memBitmap))
		win.DeleteDC(win.HDC(memhdc))

		fmt.Println(time.Since(start))
	}
}
