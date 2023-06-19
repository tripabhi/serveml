from locust import HttpUser, task

class InferenceTestUser(HttpUser):
        
    @task
    def predict(self):
       self.client.post("/predict", json={"queries" : ["Hello, how are you doing today?"]})
