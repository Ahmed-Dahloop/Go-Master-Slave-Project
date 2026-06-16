import tkinter as tk
from tkinter import scrolledtext, messagebox
import socket

HOST = '127.0.0.1'
PORT = 8080

# Function to send query
def send_query():
    query = entry.get("1.0", tk.END).strip()
    if not query:
        return
    try:
        send_btn.config(state=tk.DISABLED, text="Sending...")
        root.update()

        s = socket.socket(socket.AF_INET, socket.SOCK_STREAM)
        s.connect((HOST, PORT))

        full_message = "[GUI]\n" + query.strip() + "\n"
        s.sendall(full_message.encode())


        response = s.recv(8192).decode()
        output.delete("1.0", tk.END)
        output.insert(tk.END, response)
        s.close()
    except Exception as e:
        messagebox.showerror("Connection Error", f"Could not send query:\n{e}")
    finally:
        send_btn.config(state=tk.NORMAL, text="Send Query")

# Hover effects
def on_enter(event):
    send_btn.config(bg="#26a69a", fg="white")

def on_leave(event):
    send_btn.config(bg="#009688", fg="white")

# Root setup
root = tk.Tk()
root.title("🌟 Master DB Controller")
root.geometry("850x600")
root.configure(bg="#1e1e2f")

# Title
title = tk.Label(root, text="Master DB Controller", font=("Segoe UI", 22, "bold"), fg="#00e6ac", bg="#1e1e2f")
title.pack(pady=20)

# Query input
query_label = tk.Label(root, text="Enter SQL Query:", font=("Segoe UI", 12, "bold"), fg="white", bg="#1e1e2f")
query_label.pack(anchor="w", padx=25)

entry = tk.Text(root, height=5, width=90, font=("Consolas", 11), bg="#2e2e3f", fg="#00ffcc", insertbackground="white", bd=2, relief="flat")
entry.pack(padx=25, pady=(0, 10))

# Send button
send_btn = tk.Button(root, text="Send Query", font=("Segoe UI", 12, "bold"), bg="#009688", fg="white",
                     activebackground="#004d40", activeforeground="white", cursor="hand2", relief="flat",
                     command=send_query)
send_btn.pack(pady=10)
send_btn.bind("<Enter>", on_enter)
send_btn.bind("<Leave>", on_leave)

# Output section
output_label = tk.Label(root, text="Output:", font=("Segoe UI", 12, "bold"), fg="white", bg="#1e1e2f")
output_label.pack(anchor="w", padx=25, pady=(10, 0))

output = scrolledtext.ScrolledText(root, height=18, width=90, font=("Consolas", 11),
                                   bg="#2e2e3f", fg="white", insertbackground="white", bd=2, relief="flat")
output.pack(padx=25, pady=(0, 20))

root.mainloop()
