from fastapi import FastAPI
from pydantic import BaseModel
import typing

from app.service.inferenceservice import infer


class PredictionRequestBody(BaseModel):
    queries: list[str]


# class PredictionResponseBody(BaseModel):
#     metrics: typing.Dict[any, any]
#     predictions: list[str]


app = FastAPI()


@app.get("/")
def read_root():
    return {"Hello": "World"}


@app.post("/predict")
def run_predict(req: PredictionRequestBody):
    if len(req.queries) > 0:
        result = infer(req.queries)
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
        result = infer(req.queries)
        return {
            "metrics": {
                "TokenTime": result["TokenTime"],
                "InferenceTime": result["InferenceTime"],
                "TotalTime": result["TotalTime"],
                "ModelLoadTime": result["ModelLoadTime"],
            },
            "predictions": result["predictions"],
        }
        # predictions = infer(req.queries)
        # return {"predictions": predictions}
