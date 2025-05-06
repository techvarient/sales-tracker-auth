from http.server import HTTPServer, SimpleHTTPRequestHandler
import os
import sys
import json
import urllib.request
import urllib.parse

class CORSRequestHandler(SimpleHTTPRequestHandler):
    def end_headers(self):
        self.send_header('Access-Control-Allow-Origin', '*')
        self.send_header('Access-Control-Allow-Methods', 'GET, POST, OPTIONS')
        self.send_header('Access-Control-Allow-Headers', 'Content-Type')
        super().end_headers()
    
    def do_OPTIONS(self):
        self.send_response(200, "ok")
        self.end_headers()
    
    def do_POST(self):
        if self.path == '/auth/reset-password':
            try:
                content_length = int(self.headers.get('Content-Length', 0))
                if content_length == 0:
                    raise ValueError('Empty request body')
                    
                post_data = self.rfile.read(content_length)
                
                # Forward the request to the auth service
                req = urllib.request.Request(
                    'http://localhost:8080/auth/reset-password',
                    data=post_data,
                    headers={
                        'Content-Type': 'application/json',
                        'Content-Length': str(len(post_data))
                    },
                    method='POST'
                )
                
                with urllib.request.urlopen(req, timeout=10) as response:
                    response_data = response.read()
                    # Ensure the response is valid JSON
                    try:
                        json.loads(response_data)
                        self.send_response(response.status)
                        self.send_header('Content-type', 'application/json')
                        self.end_headers()
                        self.wfile.write(response_data)
                    except json.JSONDecodeError:
                        # If response is not JSON, wrap it in a JSON object
                        self.send_response(response.status)
                        self.send_header('Content-type', 'application/json')
                        self.end_headers()
                        self.wfile.write(json.dumps({
                            'status': 'success' if response.status < 400 else 'error',
                            'message': response_data.decode('utf-8', 'replace')
                        }).encode())
                        
            except urllib.error.HTTPError as e:
                error_body = e.read()
                try:
                    # Try to parse the error response as JSON
                    error_data = json.loads(error_body)
                    response_body = json.dumps({
                        'status': 'error',
                        'message': error_data.get('message', str(e)),
                        'code': e.code
                    }).encode()
                except:
                    # If not JSON, use the raw error body
                    response_body = json.dumps({
                        'status': 'error',
                        'message': error_body.decode('utf-8', 'replace') or str(e),
                        'code': e.code
                    }).encode()
                
                self.send_response(e.code)
                self.send_header('Content-type', 'application/json')
                self.end_headers()
                self.wfile.write(response_body)
                
            except Exception as e:
                self.send_response(500)
                self.send_header('Content-type', 'application/json')
                self.end_headers()
                self.wfile.write(json.dumps({
                    'status': 'error',
                    'message': str(e) or 'Internal server error',
                    'code': 500
                }).encode())
        else:
            self.send_response(404)
            self.end_headers()

if __name__ == '__main__':
    # Change to the directory containing reset-password.html
    os.chdir(os.path.dirname(os.path.abspath(__file__)))
    
    port = 3000
    server_address = ('', port)
    httpd = HTTPServer(server_address, CORSRequestHandler)
    
    print(f"Serving reset password page at http://localhost:{port}/reset-password.html")
    print("Press Ctrl+C to stop the server")
    
    try:
        httpd.serve_forever()
    except KeyboardInterrupt:
        print("\nServer stopped.")
        httpd.server_close()
        sys.exit(0)
