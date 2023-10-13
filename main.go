package main

import (
	"encoding/json"
	"fmt"
	"log"
	"os"
	"runtime"
	"strings"
	"syscall"
	"time"
	"unicode"
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

type Offset struct {
	DwViewMatrix           uintptr `json:"dwViewMatrix"`
	DwLocalPlayerPawn      uintptr `json:"dwLocalPlayerPawn"`
	DwEntityList           uintptr `json:"dwEntityList"`
	M_hPlayerPawn          uintptr `json:"m_hPlayerPawn"`
	M_iHealth              uintptr `json:"m_iHealth"`
	M_iTeamNum             uintptr `json:"m_iTeamNum"`
	M_vOldOrigin           uintptr `json:"m_vOldOrigin"`
	M_sSanitizedPlayerName uintptr `json:"m_sSanitizedPlayerName"`
}

var (
	user32                     = windows.NewLazySystemDLL("user32.dll")
	gdi32                      = windows.NewLazySystemDLL("gdi32.dll")
	getSystemMetrics           = user32.NewProc("GetSystemMetrics")
	setLayeredWindowAttributes = user32.NewProc("SetLayeredWindowAttributes")
	showCursor                 = user32.NewProc("ShowCursor")
	setTextAlign               = gdi32.NewProc("SetTextAlign")
	createFont                 = gdi32.NewProc("CreateFontW")
	createCompatibleDC         = gdi32.NewProc("CreateCompatibleDC")
	createSolidBrush           = gdi32.NewProc("CreateSolidBrush")
	createPen                  = gdi32.NewProc("CreatePen")
)

var teamCheck bool

func init() {
	// Ensure main() runs on the main thread.
	runtime.LockOSThread()
}

func logAndSleep(message string, err error) {
	log.Printf("%s: %v\n", message, err)
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

func getOffsets() Offset {
	var offsets Offset

	// Open the file
	offsetsJson, err := os.Open("offsets.json")
	if err != nil {
		fmt.Println("Error opening offsets.json", err)
		return offsets
	}
	defer offsetsJson.Close()

	// Decode the JSON
	err = json.NewDecoder(offsetsJson).Decode(&offsets)
	if err != nil {
		fmt.Println("Error decoding JSON:", err)
		return offsets
	}
	fmt.Println("Offset values:", offsets)
	return offsets
}

func getEntitiesInfo(procHandle windows.Handle, clientDll uintptr, screenWidth uintptr, screenHeight uintptr, offsets Offset) ([][4][2]float32, []int32, []int32, []string) {
	var entityList uintptr
	err := read(procHandle, clientDll+offsets.DwEntityList, &entityList)
	if err != nil {
		fmt.Println("Error reading initial entityList", err)
		return nil, nil, nil, nil
	}
	var (
		localPlayerP         uintptr
		localTeam            int32
		listEntry            uintptr
		entityController     uintptr
		entityControllerPawn uintptr
		entityPawn           uintptr
		entityNameAddress    uintptr
		entityTeam           int32
		entityHealth         int32
		entityName           string
		sanitizedNameStr     string
		entityOrigin         Vector3
		viewMatrix           Matrix
	)
	var (
		entRects      [][4][2]float32
		entityTeams   []int32
		entityHealths []int32
		entityNames   []string
	)
	for i := 0; i < 32; i++ {
		var sanitizedName strings.Builder
		// localPlayerP
		err = read(procHandle, clientDll+offsets.DwLocalPlayerPawn, &localPlayerP)
		if err != nil {
			fmt.Println("Error reading localPlayerP", err)
			return nil, nil, nil, nil
		}
		// listEntry
		err = read(procHandle, entityList+uintptr((8*(i&0x7FFF)>>9)+16), &listEntry)
		if err != nil {
			fmt.Println("Error reading listEntry", err)
			return nil, nil, nil, nil
		}
		if listEntry == 0 {
			continue
		}
		// entityController
		err = read(procHandle, listEntry+uintptr(120)*uintptr(i&0x1FF), &entityController)
		if err != nil {
			fmt.Println("Error reading entityController", err)
			return nil, nil, nil, nil
		}
		if entityController == 0 {
			continue
		}
		// entityControllerPawn
		err = read(procHandle, entityController+offsets.M_hPlayerPawn, &entityControllerPawn)
		if err != nil {
			fmt.Println("Error reading entityControllerPawn", err)
			return nil, nil, nil, nil
		}
		if entityControllerPawn == 0 {
			continue
		}
		// listEntry
		err = read(procHandle, entityList+uintptr(0x8*((entityControllerPawn&0x7FFF)>>9)+16), &listEntry)
		if err != nil {
			fmt.Println("Error reading listEntry", err)
			return nil, nil, nil, nil
		}
		if listEntry == 0 {
			continue
		}
		// entityPawn
		err = read(procHandle, listEntry+uintptr(120)*uintptr(entityControllerPawn&0x1FF), &entityPawn)
		if err != nil {
			fmt.Println("Error reading entityPawn", err)
			return nil, nil, nil, nil
		}
		if entityPawn == 0 {
			continue
		}
		if entityPawn == localPlayerP {
			continue
		}
		// entityTeam
		err = read(procHandle, entityPawn+offsets.M_iTeamNum, &entityTeam)
		if err != nil {
			fmt.Println("Error reading entityTeam", err)
			return nil, nil, nil, nil
		}
		if entityTeam == 0 {
			continue
		}
		if teamCheck {
			// localTeam
			err = read(procHandle, localPlayerP+offsets.M_iTeamNum, &localTeam)
			if err != nil {
				fmt.Println("Error reading localTeam", err)
				return nil, nil, nil, nil
			}
			if localTeam == entityTeam {
				continue
			}
		}
		// entityHealth
		err = read(procHandle, entityPawn+offsets.M_iHealth, &entityHealth)
		if err != nil {
			fmt.Println("Error reading entityHealth", err)
			return nil, nil, nil, nil
		}
		if entityHealth <= 0 {
			continue
		}
		// entityNameAddress
		err = read(procHandle, entityController+offsets.M_sSanitizedPlayerName, &entityNameAddress)
		if err != nil {
			fmt.Println("Error reading entityNameAddress", err)
			return nil, nil, nil, nil
		}
		// entityName
		err = read(procHandle, entityNameAddress, &entityName)
		if err != nil {
			fmt.Println("Error reading entityName", err)
			return nil, nil, nil, nil
		}
		if entityName == "" {
			continue
		}
		for _, c := range entityName {
			if unicode.IsLetter(c) || unicode.IsDigit(c) || unicode.IsPunct(c) || unicode.IsSpace(c) {
				sanitizedName.WriteRune(c)
			}
		}
		sanitizedNameStr = sanitizedName.String()
		// entityOrigin
		err = read(procHandle, entityPawn+offsets.M_vOldOrigin, &entityOrigin)
		if err != nil {
			fmt.Println("Error reading entityOrigin", err)
			return nil, nil, nil, nil
		}
		entityHead := Vector3{X: entityOrigin.X, Y: entityOrigin.Y, Z: entityOrigin.Z + 70.0}
		// viewMatrix
		err = read(procHandle, clientDll+offsets.DwViewMatrix, &viewMatrix)
		if err != nil {
			fmt.Println("Error reading viewMatrix", err)
			return nil, nil, nil, nil
		}
		screenPosHeadX, screenPosHeadY := worldToScreen(viewMatrix, entityHead)
		_, screenPosFeetY := worldToScreen(viewMatrix, entityOrigin)
		if screenPosHeadX <= -1 || screenPosFeetY <= -1 || screenPosHeadX >= float32(screenWidth) || screenPosHeadY >= float32(screenHeight) {
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
		entityNames = append(entityNames, sanitizedNameStr)
	}
	return entRects, entityTeams, entityHealths, entityNames
}

func renderEntityInfo(hdc win.HDC, tPen uintptr, gPen uintptr, oPen uintptr, rect [4][2]float32, hp int32, name string) {
	// Box
	win.SelectObject(hdc, win.HGDIOBJ(tPen))
	win.MoveToEx(hdc, int(rect[0][0]), int(rect[0][1]), nil)
	win.LineTo(hdc, int32(rect[1][0]), int32(rect[1][1]))
	win.LineTo(hdc, int32(rect[2][0]), int32(rect[2][1]))
	win.LineTo(hdc, int32(rect[3][0]), int32(rect[3][1]))
	win.LineTo(hdc, int32(rect[0][0]), int32(rect[0][1]))

	// Box outline
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

	// Health bar
	win.SelectObject(hdc, win.HGDIOBJ(gPen))
	win.MoveToEx(hdc, int(rect[0][0])-4, int(rect[3][1])+1-int(float64(int(rect[3][1])+1-int(rect[0][1]))*float64(hp)/100.0), nil)
	win.LineTo(hdc, int32(rect[3][0])-4, int32(rect[3][1])+1)

	// Health bar outline
	win.SelectObject(hdc, win.HGDIOBJ(oPen))
	win.MoveToEx(hdc, int(rect[0][0])-5, int(rect[0][1])-1, nil)
	win.LineTo(hdc, int32(rect[0][0])-5, int32(rect[3][1])+1)
	win.LineTo(hdc, int32(rect[0][0])-3, int32(rect[3][1])+1)
	win.LineTo(hdc, int32(rect[0][0])-3, int32(rect[0][1])-1)
	win.LineTo(hdc, int32(rect[0][0])-5, int32(rect[0][1])-1)

	// Name
	text, _ := windows.UTF16PtrFromString(name)
	win.SetTextColor(hdc, win.COLORREF(0xFFFFFF))
	setTextAlign.Call(uintptr(hdc), 0x00000006) // Set text alignment to center
	win.TextOut(hdc, int32(rect[0][0])+int32((int32(rect[1][0])-int32(rect[0][0]))/2), int32(rect[0][1])-14, text, int32(len(name)))
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
		logAndSleep("Error creating window class name", err)
		return 0
	}
	windowTitle, err := windows.UTF16PtrFromString("cs2go")
	if err != nil {
		logAndSleep("Error creating window title", err)
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

	if atom := win.RegisterClassEx(&wc); atom == 0 {
		logAndSleep("Error registering window class", fmt.Errorf("%v", win.GetLastError()))
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
		logAndSleep("Error creating window", fmt.Errorf("%v", win.GetLastError()))
		return 0
	}

	result, _, _ := setLayeredWindowAttributes.Call(uintptr(hwnd), 0x000000, 0, 0x00000001)
	if result == 0 {
		logAndSleep("Error setting layered window attributes", fmt.Errorf("%v", win.GetLastError()))
	}

	showCursor.Call(0)

	// Show window
	win.ShowWindow(hwnd, win.SW_SHOWDEFAULT)
	return hwnd
}

func main() {
	// take user cli input with scanner to toggle team check
	fmt.Println("Toggle team check? (Y/n)")
	var input string
	fmt.Scanln(&input)
	if input == "y" || input == "Y" {
		teamCheck = true
	} else if input == "n" || input == "N" {
		teamCheck = false
	} else {
		fmt.Println("Invalid input, defaulting to true")
		teamCheck = true
	}

	screenWidth, _, _ := getSystemMetrics.Call(0)
	screenHeight, _, _ := getSystemMetrics.Call(1)

	hwnd := initWindow(screenWidth, screenHeight)
	if hwnd == 0 {
		logAndSleep("Error creating window", fmt.Errorf("%v", win.GetLastError()))
		return
	}
	defer win.DestroyWindow(hwnd)

	// win.SetCursor()

	pid, err := findProcessId("cs2.exe")
	if err != nil {
		logAndSleep("Error finding process ID", err)
		return
	}

	clientDll, err := getModuleBaseAddress(pid, "client.dll")
	if err != nil {
		logAndSleep("Error getting client.dll base address", err)
		return
	}

	procHandle, err := getProcessHandle(pid)
	if err != nil {
		logAndSleep("Error getting process handle", err)
		return
	}

	hdc := win.GetDC(hwnd)
	if hdc == 0 {
		logAndSleep("Error getting device context", fmt.Errorf("%v", win.GetLastError()))
		return
	}

	bgBrush, _, _ := createSolidBrush.Call(uintptr(0x000000))
	if bgBrush == 0 {
		logAndSleep("Error creating brush", fmt.Errorf("%v", win.GetLastError()))
		return
	}
	defer win.DeleteObject(win.HGDIOBJ(bgBrush))
	redPen, _, _ := createPen.Call(win.PS_SOLID, 1, 0x7a78ff)
	if redPen == 0 {
		logAndSleep("Error creating pen", fmt.Errorf("%v", win.GetLastError()))
		return
	}
	defer win.DeleteObject(win.HGDIOBJ(redPen))
	greenPen, _, _ := createPen.Call(win.PS_SOLID, 1, 0x7dff78)
	if greenPen == 0 {
		logAndSleep("Error creating pen", fmt.Errorf("%v", win.GetLastError()))
		return
	}
	defer win.DeleteObject(win.HGDIOBJ(greenPen))
	bluePen, _, _ := createPen.Call(win.PS_SOLID, 1, 0xff8e78)
	if bluePen == 0 {
		logAndSleep("Error creating pen", fmt.Errorf("%v", win.GetLastError()))
		return
	}
	defer win.DeleteObject(win.HGDIOBJ(bluePen))
	outlinePen, _, _ := createPen.Call(win.PS_SOLID, 1, 0x000001)
	if outlinePen == 0 {
		logAndSleep("Error creating pen", fmt.Errorf("%v", win.GetLastError()))
		return
	}
	defer win.DeleteObject(win.HGDIOBJ(outlinePen))

	font, _, _ := createFont.Call(12, 0, 0, 0, win.FW_DONTCARE, 0, 0, 0, win.DEFAULT_CHARSET, win.OUT_DEFAULT_PRECIS, win.CLIP_DEFAULT_PRECIS, win.DEFAULT_QUALITY, win.DEFAULT_PITCH|win.FF_DONTCARE, 0)

	offsets := getOffsets()

	win.SetTimer(hwnd, 1, 15, 0)
	var msg win.MSG

	fmt.Println("Starting main loop")
	for win.GetMessage(&msg, 0, 0, 0) > 0 {
		win.TranslateMessage(&msg)
		win.DispatchMessage(&msg)

		memhdc, _, _ := createCompatibleDC.Call(uintptr(hdc))
		memBitmap := win.CreateCompatibleBitmap(hdc, int32(screenWidth), int32(screenHeight))
		win.SelectObject(win.HDC(memhdc), win.HGDIOBJ(memBitmap))
		win.SelectObject(win.HDC(memhdc), win.HGDIOBJ(bgBrush))
		win.SetBkMode(win.HDC(memhdc), win.TRANSPARENT)
		win.SelectObject(win.HDC(memhdc), win.HGDIOBJ(font))

		rects, teams, healths, names := getEntitiesInfo(procHandle, clientDll, screenWidth, screenHeight, offsets)
		for i, rect := range rects {
			if teams[i] == 2 {
				renderEntityInfo(win.HDC(memhdc), redPen, greenPen, outlinePen, rect, healths[i], names[i])
			} else {
				renderEntityInfo(win.HDC(memhdc), bluePen, greenPen, outlinePen, rect, healths[i], names[i])
			}
		}
		win.BitBlt(hdc, 0, 0, int32(screenWidth), int32(screenHeight), win.HDC(memhdc), 0, 0, win.SRCCOPY)

		// Delete the memory bitmap and device context
		win.DeleteObject(win.HGDIOBJ(memBitmap))
		win.DeleteDC(win.HDC(memhdc))
	}
}
