from typing import Union

from fastapi import APIRouter, Request, HTTPException

from tracking.models import ChirpstackBaseEventModel, ChirpstackUpEventModel

gateway_router = APIRouter(
    prefix="/gateway", tags=["gateway"])

@gateway_router.post("/data", summary="Endpoint for Chirpstack HTTP integration")
async def up_event(request: Request, event: str, data: Union[ChirpstackUpEventModel, ChirpstackBaseEventModel]):

    if event != "up":
        # Log all other event types
        print(f"Event {event} not implemented!")
        return {"status": "success"}
    elif event == "up" and not isinstance(data, ChirpstackUpEventModel):
        raise HTTPException(status_code=404, detail="Wrong model for event type")


    print(f"Event {event}!")
    print(data)

    print(await request.body())
    return {"status": "success"}
