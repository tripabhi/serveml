from flask import Flask, request
from gevent.pywsgi import WSGIServer

from service.inferenceservice import InferenceService

app = Flask(__name__)
svc = InferenceService()


@app.route("/")
def read_root():
    return {"Hello": "World"}


@app.route("/predict", methods=["POST"])
def run_predict():
    data = request.json
    if len(data["queries"]) > 0:
        result = svc.infer(data["queries"])
        return {
            "metrics": {
                "InferenceTime": result["InferenceTime"],
                "TotalTime": result["TotalTime"],
            },
            "predictions": result["predictions"],
        }


@app.route("/predict_no_batch", methods=["POST"])
def run_predict_no_batch():
    data = request.json
    if len(data["queries"]) > 0:
        result = svc.infer(data["queries"])
        return {
            "metrics": {
                "InferenceTime": result["InferenceTime"],
                "TotalTime": result["TotalTime"],
            },
            "predictions": result["predictions"],
        }


if __name__ == "__main__":
    try:
        server = WSGIServer(("0.0.0.0", 3000), app)
        print("Starting Inference Function Server on 0.0.0.0:3000")
        server.serve_forever()
    except Exception as e:
        print("Exception occured : ", e)
    finally:
        server.stop()
