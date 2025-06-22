from typing import Optional

from pydantic import Field
from pydantic_settings import BaseSettings


class Settings(BaseSettings):
    mcp_enabled: bool = Field(default=False)
    mcp_url: Optional[str] = Field(default=None)
    mcp_api_key: Optional[str] = Field(default=None)
    # Update interval in seconds
    mcp_update_interval: int = Field(default=60, ge=5)
    