from locust import FastHttpUser, task


class InferenceTestUser(FastHttpUser):
    @task
    def predict(self):
        self.client.post(
            "/predict_no_batch", json={"queries": ["Hello, how are you doing today?"]}
        )
