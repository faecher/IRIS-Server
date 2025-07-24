from tracking.db import SessionLocal


async def get_db():
    database = SessionLocal()
    try:
        yield database
    finally:
        database.close()
