from fastapi import FastAPI
from pydantic import BaseModel

from app.service.inferenceservice import InferenceService


class PredictionRequestBody(BaseModel):
    queries: list[str]


svc = InferenceService()
app = FastAPI()


@app.get("/")
def read_root():
    return {"Hello": "World"}


@app.post("/predict")
def run_predict(req: PredictionRequestBody):
    if len(req.queries) > 0:
        result = svc.infer(req.queries)
        return {
            "metrics": {
                "TokenTime": result["TokenTime"],
                "InferenceTime": result["InferenceTime"],
                "TotalTime": result["TotalTime"],
                "ModelLoadTime": result["ModelLoadTime"],
            },
            "predictions": result["predictions"],
        }


@app.post("/predictNoBatcher")
def run_predict(req: PredictionRequestBody):
    if len(req.queries) > 0:
        result = svc.infer(req.queries)
        return {
            "metrics": {
                "TokenTime": result["TokenTime"],
                "InferenceTime": result["InferenceTime"],
                "TotalTime": result["TotalTime"],
                "ModelLoadTime": result["ModelLoadTime"],
            },
            "predictions": result["predictions"],
        }
