import asyncio

import socketio

socket = socketio.AsyncServer(async_mode="asgi", cors_allowed_origins="*")


@socket.event
async def connect(sid, environ):
    print("A new client connected!", sid)
    await socket.emit('update', {"devices": [
        {"name": "Test", "lat": 8, "long": 49, "status": 6}
    ]})


@socket.event
def disconnect(sid):
    print('A client disconnected!', sid)


@socket.on("register")
def register_event(sid, data):
    print("Register event")


@socket.on("updated_view")
def on_view_updated(sid, data):
    print("update_view event")


@socket.on("request_display")
def request_display_event(sid, data):
    print("Received request_display event")


@socket.on("request_display_for_all")
def request_display_event_all(sid, data):
    print("Received request_display_for_all event")
