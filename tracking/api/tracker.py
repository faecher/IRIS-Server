from fastapi import APIRouter

tracker_router = APIRouter(
    prefix="/tracker", tags=["tracker"])


@tracker_router.get("/")
def get_tracker():
    pass
