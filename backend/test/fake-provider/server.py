import json
from http.server import BaseHTTPRequestHandler, HTTPServer
from urllib.parse import urlparse

class Handler(BaseHTTPRequestHandler):
    def do_GET(self):
        path = urlparse(self.path).path
        if path.startswith('/v3/place/'):
            self.reply({'status':'1','info':'OK','pois':[{'id':'fixture-toilet-1','name':'测试公共厕所','typecode':'200300','address':'测试路1号','location':'116.397470,39.908823','citycode':'010','adcode':'110101'}]})
            return
        if path.startswith('/v3/direction/') or path.startswith('/v4/direction/'):
            self.reply({'status':'1','route':{'paths':[{'distance':'1200','duration':'900','polyline':'116.397470,39.908823;116.407470,39.918823','steps':[{'instruction':'向前行驶','distance':'1200','duration':'900','polyline':'116.397470,39.908823;116.407470,39.918823'}]}]},'data':{'paths':[]}})
            return
        self.send_error(404)
    def log_message(self, *_):
        pass
    def reply(self, value):
        raw=json.dumps(value,ensure_ascii=False).encode()
        self.send_response(200);self.send_header('Content-Type','application/json');self.send_header('Content-Length',str(len(raw)));self.end_headers();self.wfile.write(raw)

HTTPServer(('0.0.0.0',8090),Handler).serve_forever()
