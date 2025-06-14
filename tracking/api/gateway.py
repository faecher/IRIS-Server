from typing import Union

from fastapi import APIRouter, HTTPException, Depends
from sqlalchemy.orm import Session

from tracking.db.crud import get_tracker_by_eui, create_tracker, update_tracker
from tracking.dependencies import get_db
from tracking.models import ChirpstackBaseEventModel, ChirpstackUpEventModel

gateway_router = APIRouter(
    prefix="/gateway", tags=["gateway"])


@gateway_router.post("/data", summary="Endpoint for Chirpstack HTTP integration")
async def up_event(event: str, data: Union[ChirpstackUpEventModel, ChirpstackBaseEventModel],
                   db: Session = Depends(get_db)):
    if event != "up":
        # Log all other event types
        print(f"Event {event} not implemented, skipping it!")
        return {"status": "success"}
    elif event == "up" and not isinstance(data, ChirpstackUpEventModel):
        print("Unknown model for up event!")
        raise HTTPException(status_code=404, detail="Wrong model for event type")

    # Process the request data
    print(f"Event {event}!")
    print(f"Device {data.deviceInfo.devEui}")

    tracker = get_tracker_by_eui(db, data.deviceInfo.devEui)

    if tracker is None:
        print("No tracker found, creating a new tracker")
        create_tracker(db, data)
    else:
        print("Tracker found, updating data")
        update_tracker(db, data)

    return {"status": "success"}
