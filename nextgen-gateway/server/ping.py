import json
import threading
from http.server import BaseHTTPRequestHandler, HTTPServer


class SimpleHandler(BaseHTTPRequestHandler):
    def log_message(self, format, *args):
        return

    def do_GET(self):
        if self.path == "/ping":
            self.send_response(200)
            self.send_header("Content-Type", "application/json")
            self.end_headers()
            self.wfile.write(json.dumps({"message": "pong"}).encode("utf-8"))
        else:
            self.send_response(404)
            self.end_headers()


def serve(port: int):
    server = HTTPServer(("127.0.0.1", port), SimpleHandler)
    print(f"Server running on http://127.0.0.1:{port}")
    server.serve_forever()


if __name__ == "__main__":
    ports = [8081, 8082, 8083]
    threads = []

    for p in ports:
        t = threading.Thread(target=serve, args=(p,), daemon=True)
        t.start()
        threads.append(t)

    try:
        for t in threads:
            t.join()
    except KeyboardInterrupt:
        print("\nShutting down...")

