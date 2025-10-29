from fastapi import FastAPI
from . import routes

def create_app():
    app = FastAPI()
    app.include_router(routes.router)
    return app
