from tracking import Settings


async def get_mcp_data():
    settings = Settings()

    if settings.mcp_enabled is True and settings.mcp_url is not None:
        print("Getting MCP data :O")
    else:
        print("Skipping MPC query due to missing configuration")
