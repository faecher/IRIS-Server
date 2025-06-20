import socketio
from fastapi import FastAPI
from fastapi.middleware.cors import CORSMiddleware

from tracking.settings import Settings
from tracking.api import api_router
from tracking.db import models, engine
from tracking.socket import socket
from tracking.tasks import get_mcp_data
from tracking.utils.scheduling import repeat_every

__version__ = "1.0.0"

# Create all models from
models.Base.metadata.create_all(bind=engine)

# Create a settings instance
settings = Settings()

app = FastAPI(
    summary="Backend server for display software",
    version=__version__
)

# app.add_middleware(
#    CORSMiddleware,
#    allow_origins=["*"],
#    allow_credentials=True,
#    allow_methods=["*"],
#    allow_headers=["*"],
# )

# Create a combined instance of FastAPI and Socket.IO. If running the application, always use the socket_app or else
# there will be issues with the registering mounts and receiving certain request types.
socket_app = socketio.ASGIApp(socketio_server=socket, other_asgi_app=app)

# Include all routers
app.include_router(api_router)


@app.get("/")
async def root():
    return {"message": "Hello World"}


@app.on_event("startup")
@repeat_every(seconds=settings.mcp_update_interval, raise_exceptions=True)
async def schedule_lookup():
    await get_mcp_data()
