from locust import FastHttpUser, task, between


class InferenceTestUser(FastHttpUser):
    @task
    def predict(self):
        self.client.post(
            "/predictions/benchmark",
            json={
                "text": "Bloomberg has decided to publish a new report on the global economy.",
                "target": 1,
            },
        )
