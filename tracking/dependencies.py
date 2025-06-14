from tracking.db import SessionLocal


def get_db():
    database = SessionLocal()
    try:
        yield database
    finally:
        database.close()
