import tkinter as tk
from tkinter import scrolledtext, messagebox
import socket

HOST = '192.168.1.13'  # IP of master
PORT = 8080

def send_query():
    query = entry.get("1.0", tk.END).strip()
    if not query:
        return

    try:
        send_button.config(state=tk.DISABLED, text="Sending...")
        root.update()
        
        s = socket.socket(socket.AF_INET, socket.SOCK_STREAM)
        s.connect((HOST, PORT))
        s.sendall(query.encode())
        response = s.recv(8192).decode()
        output.delete("1.0", tk.END)
        output.insert(tk.END, response)
        s.close()
    except Exception as e:
        messagebox.showerror("Error", f"Could not send query:\n{e}")
    finally:
        send_button.config(state=tk.NORMAL, text="Send Query")

def on_enter(e):
    send_button.config(bg="#009688", fg="white")

def on_leave(e):
    send_button.config(bg="#00bfa5", fg="white")

root = tk.Tk()
root.title("💻 Slave DB Interface")
root.geometry("800x600")
root.configure(bg="#1e1e2f")

# Title
title_label = tk.Label(root, text="Slave Database Client", font=("Helvetica", 20, "bold"),
                       bg="#1e1e2f", fg="#00e6ac")
title_label.pack(pady=20)

# Query Label
query_label = tk.Label(root, text="Enter SQL Query:", font=("Helvetica", 12, "bold"),
                       bg="#1e1e2f", fg="white")
query_label.pack(anchor="w", padx=20)

# Query Entry Box
entry = tk.Text(root, height=5, width=80, font=("Consolas", 11), bg="#2e2e3f", fg="#00ffcc", insertbackground="white")
entry.pack(padx=20, pady=(0, 10))

# Send Button
send_button = tk.Button(root, text="Send Query", font=("Helvetica", 12, "bold"),
                        bg="#00bfa5", fg="white", activebackground="#009688",
                        cursor="hand2", command=send_query)
send_button.pack(pady=10)
send_button.bind("<Enter>", on_enter)
send_button.bind("<Leave>", on_leave)

# Output Label
output_label = tk.Label(root, text="Output:", font=("Helvetica", 12, "bold"),
                        bg="#1e1e2f", fg="white")
output_label.pack(anchor="w", padx=20, pady=(10, 0))

# Output Box
output = scrolledtext.ScrolledText(root, height=18, width=80, font=("Consolas", 11),
                                   bg="#2e2e3f", fg="#ffffff", insertbackground="white")
output.pack(padx=20, pady=(0, 20))

root.mainloop()
