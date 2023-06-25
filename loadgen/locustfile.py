from locust import FastHttpUser, task


class InferenceTestUser(FastHttpUser):
    @task
    def predict(self):
        self.client.post(
            "/predictions/bert_sa",
            json={
                "text": "Bloomberg has decided to publish a new report on the global economy.",
                "target": 1,
            },
        )
