from fastapi import FastAPI
from pydantic import BaseModel

from service.inferenceservice import infer

class PredictionRequestBody(BaseModel):
    queries: list[str]
    
    
class PredictionResponseBody(BaseModel):
    predictions: list[str]

app = FastAPI()

@app.get("/")
def read_root():
    return {"Hello" : "World"}

@app.post("/predict")
def run_predict(req: PredictionRequestBody) -> PredictionResponseBody :
    if len(req.queries) > 0 :
        predictions = infer(req.queries)
        return { predictions }
        