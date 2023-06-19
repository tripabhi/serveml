from locust import task, FastHttpUser


class InferenceTestUser(FastHttpUser):
    @task
    def predict(self):
        with self.rest(
            "POST",
            "http://10.152.183.147/predict",
            json={"queries": ["Hello, how are you doing today?"]},
        ) as resp:
            resp.success(f"Successful")
