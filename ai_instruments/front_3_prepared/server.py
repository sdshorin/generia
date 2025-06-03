#!/usr/bin/env python3
"""
Simple HTTP server for serving static files
Usage: python3 server.py [port]
Default port: 8000
"""

import http.server
import socketserver
import os
import sys
import webbrowser
from urllib.parse import urlparse

class CustomHTTPRequestHandler(http.server.SimpleHTTPRequestHandler):
    """Custom handler to serve index.html for directory requests"""
    
    def end_headers(self):
        # Add CORS headers for local development
        self.send_header('Access-Control-Allow-Origin', '*')
        self.send_header('Access-Control-Allow-Methods', 'GET, POST, OPTIONS')
        self.send_header('Access-Control-Allow-Headers', 'Content-Type')
        super().end_headers()
    
    def do_GET(self):
        """Handle GET requests"""
        parsed_path = urlparse(self.path)
        path = parsed_path.path
        
        # Redirect root to main.html
        if path == '/':
            self.send_response(302)
            self.send_header('Location', '/main.html')
            self.end_headers()
            return
        
        # Handle directory requests
        if path.endswith('/'):
            path += 'main.html'
            self.path = path
        
        # Serve the file
        super().do_GET()

def main():
    """Main function to start the server"""
    # Get port from command line argument or use default
    port = 8000
    if len(sys.argv) > 1:
        try:
            port = int(sys.argv[1])
        except ValueError:
            print(f"Invalid port number: {sys.argv[1]}")
            sys.exit(1)
    
    # Change to the script directory
    script_dir = os.path.dirname(os.path.abspath(__file__))
    os.chdir(script_dir)
    
    # Create server
    with socketserver.TCPServer(("", port), CustomHTTPRequestHandler) as httpd:
        server_url = f"http://localhost:{port}"
        
        print("=" * 60)
        print("🌟 GENERIA STATIC SERVER")
        print("=" * 60)
        print(f"📍 Server running at: {server_url}")
        print(f"📁 Serving files from: {script_dir}")
        print("")
        print("📄 Available pages:")
        print(f"   • Main page:      {server_url}/main.html")
        print(f"   • Catalog:        {server_url}/catalog.html")
        print(f"   • Create World:   {server_url}/create-world.html")
        print(f"   • World Feed:     {server_url}/world-feed.html")
        print(f"   • Post Detail:    {server_url}/post-detail.html")
        print(f"   • World About:    {server_url}/world-about.html")
        print(f"   • Character:      {server_url}/character-profile.html")
        print(f"   • Settings:       {server_url}/settings.html")
        print(f"   • Login:          {server_url}/login.html")
        print(f"   • Register:       {server_url}/register.html")
        print("")
        print("⚠️  Note: This is a development server only!")
        print("💡 Press Ctrl+C to stop the server")
        print("=" * 60)
        
        # Open browser automatically
        try:
            webbrowser.open(server_url)
            print("🌐 Opening browser automatically...")
        except Exception as e:
            print(f"Could not open browser automatically: {e}")
        
        print("")
        
        try:
            httpd.serve_forever()
        except KeyboardInterrupt:
            print("\n🛑 Server stopped by user")
            print("👋 Goodbye!")

if __name__ == "__main__":
    main()