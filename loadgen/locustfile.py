from locust import FastHttpUser, task


class InferenceTestUser(FastHttpUser):
    @task
    def predict(self):
        self.client.post(
            "/predict", json={"queries": ["Hello, how are you doing today?"]}
        )
