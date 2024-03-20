import os
import subprocess
import json
import re
import urllib.parse
from http.server import BaseHTTPRequestHandler, HTTPServer

ROMS_DIR = "/roms/ports/amberserver/"
DOWNLOAD_LOG = "/roms/ports/amberserver/download.log"
IA_URL = "https://archive.org/download/ia-pex/ia"


class DownloadServer(BaseHTTPRequestHandler):
    def do_GET(self):
        if self.path =='/':
            self.send_response(200)
            self.send_header("Content-type", "text/html")
            self.end_headers()
            with open("static/index.html") as f:
                self.wfile.write(f.read().encode())
        elif self.path.startswith("/platforms/"):
            self.handle_platforms()
        elif self.path.startswith("/download/"):
            self.handle_download()
        elif self.path.startswith("/progress"):
            self.handle_progress()
        else:
            self.send_response(404)
            self.end_headers()
            self.wfile.write(b"Not Found")

    def handle_platforms(self):
        platform_id = self.path[len("/platforms/"):]
        ia_command = [os.path.join(ROMS_DIR, "ia"), "list", platform_id]
        try:
            output = subprocess.check_output(ia_command).decode("utf-8")
            game_list = []
            for line in output.split("\n"):
                if any(line.endswith(ext) for ext in (".cue", ".bin", ".7z", ".zip", ".chd", ".pbp", ".cdi")):
                    filename = os.path.basename(line)
                    name = os.path.splitext(filename)[0]
                    game_list.append({"nome": name, "arquivo": line})
            self.send_json_response(game_list)
        except subprocess.CalledProcessError as e:
            self.send_error(500, f"Failed to execute command: {e}")

    def handle_download(self):
        _, _, query = self.path.partition("?")
        query_dict = urllib.parse.parse_qs(query)
        platform = self.path.split("/")[2]  # Alteração aqui para pegar a plataforma diretamente
        arquivo = query_dict.get("arquivo", [""])[0].replace("%20", " ")

        # Separando o nome do arquivo em folder e arquivo
        folder, arquivo = arquivo.split("/")

        # subprocess.run(["echo", "0%", ">", DOWNLOAD_LOG], check=True)
        downloader_script = os.path.join(ROMS_DIR, "ia")
        # subprocess.run([downloader_script, platform, folder, arquivo], check=True)
        print(f'Plataform: {platform}')
        print(f'arquivo: {arquivo}')
        print(f'folder: {folder}')
        self.send_response(200)
        self.send_header("Content-type", "text/plain")
        self.end_headers()
        self.wfile.write(b"Arquivo baixado com sucesso!")

    def handle_progress(self):
        if os.path.exists(DOWNLOAD_LOG):
            with open(DOWNLOAD_LOG, "r") as f:
                content = f.read()
                matches = re.findall(r"\d+%", content)
                highest_percent = max(map(int, matches)) if matches else 0
                self.send_response(200)
                self.send_header("Content-type", "text/plain")
                self.end_headers()
                self.wfile.write(f"{highest_percent}%".encode())
        else:
            self.send_response(200)
            self.send_header("Content-type", "text/plain")
            self.end_headers()
            self.wfile.write(b"0%")

    def send_json_response(self, data):
        self.send_response(200)
        self.send_header("Content-type", "application/json")
        self.end_headers()
        self.wfile.write(json.dumps(data).encode())


def download_ia():
    ia_path = os.path.join(ROMS_DIR, "ia")
    if not os.path.exists(ia_path):
        print("Downloading ia executable...")
        subprocess.run(["wget", "-O", ia_path, IA_URL], check=True)
        os.chmod(ia_path, 0o755)


def main():
    download_ia()
    server_address = ("", 8080)
    httpd = HTTPServer(server_address, DownloadServer)
    print("Server running at localhost:8080")
    httpd.serve_forever()


if __name__ == "__main__":
    main()
