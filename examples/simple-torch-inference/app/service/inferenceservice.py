from transformers import AutoTokenizer, AutoModelForSequenceClassification
import torch
import os
import pathutil
import typing

output_map = {1: 'Negative', 2: 'Neutral', 3: 'Positive', 4: 'Very Positive', 5: 'Very Negative'}

_data_path = pathutil.get_data_path()

modelname = os.path.join(_data_path, 'model_sentiment.pt')
tokenname = os.path.join(_data_path, 'tokenizer_sentiment.pt')

model = AutoModelForSequenceClassification.from_pretrained(modelname)
tokenizer = AutoTokenizer.from_pretrained(tokenname)

def infer(data_list: typing.List[str]):
    tokens = tokenizer.batch_encode_plus(data_list, 
                                         padding=True, 
                                         truncation=True, 
                                         return_tensors='pt')
    
    results = model(**tokens)
    predictions = torch.argmax(results.logits, dim = 1) + 1
    sentiment_labels = [output_map[p.item()] for p in predictions]
    return sentiment_labels