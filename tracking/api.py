from fastapi import APIRouter

api_router = APIRouter(
    prefix="/api"
)

@api_router.get("/", summary="Get lastest blog post")
async def get_blog_post():
    return {"test": 123}

