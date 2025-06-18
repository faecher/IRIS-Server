import asyncio

import socketio

from tracking.dependencies import get_db
from tracking.db.models import Tracker
from tracking.models import TrackerModel

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


@socket.on("requestTrackerData")
async def request_tracker_data(sid, data):
    print("Requesting tracker data")

    # Get the database
    db = next(get_db())

    trackers = db.query(Tracker).all()
    # Convert all sqlalchemy objects to pydantic models for serialization
    response = []
    for tracker in trackers:
        a = TrackerModel.model_validate(tracker).model_dump()
        response.append(a)

    await socket.emit('getTrackerData', {"devices": response})


@socket.on("updated_view")
def on_view_updated(sid, data):
    print("update_view event")


@socket.on("request_display")
def request_display_event(sid, data):
    print("Received request_display event")


@socket.on("request_display_for_all")
def request_display_event_all(sid, data):
    print("Received request_display_for_all event")
