package main

import (
	"fmt"
	"net/http"
	"os"
	"time"
)

// The entire MyOS frontend is embedded as a raw string.
// This prevents Node.js-style syntax errors by serving it as a static response.
const indexHTML = `
<!DOCTYPE html>
<html lang="en">
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <title>MyOS - Go Edition</title>
    <!-- Firebase Compat SDKs -->
    <script src="https://www.gstatic.com/firebasejs/10.7.1/firebase-app-compat.js"></script>
    <script src="https://www.gstatic.com/firebasejs/10.7.1/firebase-auth-compat.js"></script>
    <script src="https://www.gstatic.com/firebasejs/10.7.1/firebase-firestore-compat.js"></script>
    <style>
        :root {
            --win-blue: #0078d4;
            --taskbar-bg: rgba(28, 28, 28, 0.85);
            --text-main: #ffffff;
        }

        body, html {
            margin: 0; padding: 0; width: 100%; height: 100%;
            font-family: "Segoe UI", "Tahoma", sans-serif;
            background-color: #000; color: var(--text-main);
            overflow: hidden; user-select: none;
        }

        /* --- Boot Animation --- */
        #boot-screen {
            position: fixed; inset: 0; background: #000;
            display: flex; flex-direction: column; align-items: center; justify-content: center;
            z-index: 1000; transition: opacity 1s ease;
        }

        .logo {
            width: 100px; height: 100px; margin-bottom: 60px;
            display: grid; grid-template-columns: 1fr 1fr; gap: 5px;
        }
        .logo-part { background-color: var(--win-blue); width: 48px; height: 48px; }

        .loader {
            width: 40px; height: 40px; border: 3px solid rgba(255,255,255,0.1);
            border-top-color: #fff; border-radius: 50%;
            animation: spin 1s linear infinite;
        }
        @keyframes spin { to { transform: rotate(360deg); } }

        /* --- Desktop Layout --- */
        #desktop {
            position: fixed; inset: 0;
            background: url('https://images.unsplash.com/photo-1633533412356-6e4653556094?q=80&w=2070&auto=format&fit=crop') center/cover;
            display: none;
        }

        #taskbar {
            position: fixed; bottom: 0; width: 100%; height: 48px;
            background: var(--taskbar-bg); backdrop-filter: blur(20px);
            display: flex; align-items: center; justify-content: space-between;
            padding: 0 15px; box-sizing: border-box;
            border-top: 1px solid rgba(255,255,255,0.1);
            z-index: 900;
        }

        .start-btn {
            width: 36px; height: 36px; display: flex; align-items: center; justify-content: center;
            border-radius: 4px; cursor: pointer; transition: 0.2s;
        }
        .start-btn:hover { background: rgba(255,255,255,0.1); }

        #start-menu {
            position: fixed; bottom: 55px; left: 10px; width: 420px; height: 550px;
            background: rgba(32, 32, 32, 0.95); backdrop-filter: blur(30px);
            border-radius: 8px; border: 1px solid rgba(255,255,255,0.1);
            transform: translateY(120%); transition: transform 0.4s cubic-bezier(0.1, 0.9, 0.2, 1);
            padding: 25px; box-sizing: border-box; z-index: 850;
        }
        #start-menu.open { transform: translateY(0); }

        .tray { font-size: 12px; text-align: right; line-height: 1.2; }
    </style>
</head>
<body>

    <div id="boot-screen">
        <div class="logo">
            <div class="logo-part"></div><div class="logo-part"></div>
            <div class="logo-part"></div><div class="logo-part"></div>
        </div>
        <div class="loader"></div>
        <p style="margin-top: 25px; font-weight: 300; letter-spacing: 1px;">Starting Go-OS Engine...</p>
    </div>

    <div id="desktop">
        <div id="start-menu">
            <h2 style="font-weight: 400; margin-top: 0;">MyOS</h2>
            <p style="opacity: 0.7; font-size: 14px;">Welcome to your desktop environment powered by Go.</p>
            <div style="margin-top: auto; border-top: 1px solid rgba(255,255,255,0.1); padding-top: 20px;">
                <button onclick="location.reload()" style="background: rgba(255,255,255,0.1); border: none; color: white; padding: 8px 15px; border-radius: 4px; cursor: pointer;">Restart System</button>
            </div>
        </div>

        <div id="taskbar">
            <div class="start-btn" id="startBtn">
                <svg width="24" height="24" viewBox="0 0 24 24" fill="#00ADEF">
                    <rect x="3" y="3" width="8" height="8"/><rect x="13" y="3" width="8" height="8"/>
                    <rect x="3" y="13" width="8" height="8"/><rect x="13" y="13" width="8" height="8"/>
                </svg>
            </div>
            
            <div class="tray">
                <div id="clock-time" style="font-weight: 500;">12:00 PM</div>
                <div id="clock-date" style="opacity: 0.6; font-size: 10px;">01/01/2026</div>
            </div>
        </div>
    </div>

    <script>
        // Boot Logic
        window.onload = () => {
            setTimeout(() => {
                const boot = document.getElementById('boot-screen');
                boot.style.opacity = '0';
                setTimeout(() => {
                    boot.style.display = 'none';
                    document.getElementById('desktop').style.display = 'block';
                }, 1000);
            }, 3000);
        };

        // UI Interactions
        const startBtn = document.getElementById('startBtn');
        const startMenu = document.getElementById('start-menu');

        startBtn.addEventListener('click', (e) => {
            e.stopPropagation();
            startMenu.classList.toggle('open');
        });

        document.addEventListener('click', (e) => {
            if (!startMenu.contains(e.target)) {
                startMenu.classList.remove('open');
            }
        });

        // System Clock
        function updateClock() {
            const now = new Date();
            document.getElementById('clock-time').innerText = now.toLocaleTimeString([], { hour: '2-digit', minute: '2-digit' });
            document.getElementById('clock-date').innerText = now.toLocaleDateString();
        }
        setInterval(updateClock, 1000);
        updateClock();
    </script>
</body>
</html>
`

func main() {
	// 1. Determine Port (Render uses the PORT env var)
	port := os.Getenv("PORT")
	if port == "" {
		port = "3000"
	}

	// 2. Setup the Router
	mux := http.NewServeMux()

	// 3. Serve the indexHTML for all root requests
	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		w.Header().Set("X-Backend-Server", "Golang-Net-Http")
		fmt.Fprint(w, indexHTML)
	})

	// 4. Configure Server with timeouts
	srv := &http.Server{
		Addr:         "0.0.0.0:" + port,
		Handler:      mux,
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 15 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	fmt.Printf("==> MyOS Go Server starting on port %s\n", port)
	
	// 5. Start Listening
	if err := srv.ListenAndServe(); err != nil {
		fmt.Fprintf(os.Stderr, "Error starting server: %v\n", err)
		os.Exit(1)
	}
}
