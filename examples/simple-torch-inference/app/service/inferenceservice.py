from transformers import AutoTokenizer, AutoModelForSequenceClassification
import torch
import os
import app.pathutil
import time
import typing

output_map = {
    1: "Negative",
    2: "Neutral",
    3: "Positive",
    4: "Very Positive",
    5: "Very Negative",
}

_data_path = app.pathutil.get_data_path()

time_model_load_start = time.monotonic()
modelname = os.path.join(_data_path, "model_sentiment.pt")
tokenname = os.path.join(_data_path, "tokenizer_sentiment.pt")
time_model_load_end = time.monotonic()

model = AutoModelForSequenceClassification.from_pretrained(modelname)
tokenizer = AutoTokenizer.from_pretrained(tokenname)


def convert_to_ms(num):
    return int(round(num * 1000))


def infer(data_list: typing.List[str]):
    t_start = time.monotonic()
    tokens = tokenizer.batch_encode_plus(
        data_list, padding=True, truncation=True, return_tensors="pt"
    )
    t_token = time.monotonic()

    results = model(**tokens)
    t_inference = time.monotonic()
    value = {}
    value["TokenTime"] = convert_to_ms(t_token - t_start)
    value["InferenceTime"] = convert_to_ms(t_inference - t_token)
    value["TotalTime"] = convert_to_ms(t_inference - t_start)
    value["ModelLoadTime"] = convert_to_ms(time_model_load_end - time_model_load_start)

    predictions = torch.argmax(results.logits, dim=1) + 1
    sentiment_labels = [output_map[p.item()] for p in predictions]
    value["predictions"] = sentiment_labels
    return value
