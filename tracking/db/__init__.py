from sqlalchemy import Column as Col
from sqlalchemy import create_engine
from sqlalchemy.orm import sessionmaker
from sqlalchemy.ext.declarative import declarative_base
from functools import partial


from tracking import Settings


settings = Settings()

# Load the database uri from settings
SQLALCHEMY_DATABASE_URL = settings.sqlalchemy_database_url

CONNECTION_ARGS = {}
# Add additional options for sqlite databases
if SQLALCHEMY_DATABASE_URL.startswith("sqlite"):
    CONNECTION_ARGS = {"check_same_thread": False}

# Create a new database engine
engine = create_engine(
    SQLALCHEMY_DATABASE_URL, connect_args=CONNECTION_ARGS
)
# Create a session from the database engine
SessionLocal = sessionmaker(autocommit=False, autoflush=False, bind=engine)
# Get the base class for our models
Base = declarative_base()

# SQLAlchemy treats columns as nullable by default, which we don't want.
Column = partial(Col, nullable=False)
