from transformers import AutoTokenizer, AutoModelForSequenceClassification
import torch
import os
import app.pathutil
import time
import typing


class InferenceService:
    def __init__(self):
        __data_path = app.pathutil.get_data_path()
        __modelname = os.path.join(__data_path, "model_sentiment.pt")
        __tokenname = os.path.join(__data_path, "tokenizer_sentiment.pt")
        self.model = AutoModelForSequenceClassification.from_pretrained(__modelname)
        self.tokenizer = AutoTokenizer.from_pretrained(__tokenname)
        self.output_map = {
            1: "Negative",
            2: "Neutral",
            3: "Positive",
            4: "Very Positive",
            5: "Very Negative",
        }

    def convert_to_ms(self, num):
        return int(round(num * 1000))

    def infer(self, data_list: typing.List[str]):
        t_start = time.monotonic()
        tokens = self.tokenizer.batch_encode_plus(
            data_list, padding=True, truncation=True, return_tensors="pt"
        )
        t_token = time.monotonic()

        results = self.model(**tokens)
        t_inference = time.monotonic()
        value = {}
        value["InferenceTime"] = self.convert_to_ms(t_inference - t_token)
        value["TotalTime"] = self.convert_to_ms(t_inference - t_start)

        predictions = torch.argmax(results.logits, dim=1) + 1
        sentiment_labels = [self.output_map[p.item()] for p in predictions]
        value["predictions"] = sentiment_labels
        return value
