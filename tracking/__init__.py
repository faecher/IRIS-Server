import socketio
from fastapi import FastAPI
from fastapi.middleware.cors import CORSMiddleware

from tracking.api import api_router
from tracking.db import models, engine
from tracking.socket import socket

__version__ = "1.0.0"

# Create all models from
models.Base.metadata.create_all(bind=engine)

app = FastAPI(
    summary="Backend server for display software",
    version=__version__
)

#app.add_middleware(
#    CORSMiddleware,
#    allow_origins=["*"],
#    allow_credentials=True,
#    allow_methods=["*"],
#    allow_headers=["*"],
#)

# Create a combined instance of FastAPI and Socket.IO. If running the application, always use the socket_app or else
# there will be issues with the registering mounts and receiving certain request types.
socket_app = socketio.ASGIApp(socketio_server=socket, other_asgi_app=app)

# Include all routers
app.include_router(api_router)

@app.get("/")
async def root():
    return {"message": "Hello World"}

