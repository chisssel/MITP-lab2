import pytest
from fastapi.testclient import TestClient
from main import app

client = TestClient(app)


class TestPingEndpoint:
    def test_ping_returns_200(self):
        response = client.get("/ping")
        assert response.status_code == 200

    def test_ping_returns_correct_message(self):
        response = client.get("/ping")
        assert response.json() == {"message": "pong"}

    def test_ping_content_type(self):
        response = client.get("/ping")
        assert response.headers["content-type"] == "application/json"


class TestJsonEndpoint:
    def test_json_returns_200(self):
        response = client.get("/json")
        assert response.status_code == 200

    def test_json_has_required_keys(self):
        response = client.get("/json")
        data = response.json()
        assert "status" in data
        assert "timestamp" in data
        assert "id" in data
        assert "data" in data

    def test_json_status_is_ok(self):
        response = client.get("/json")
        assert response.json()["status"] == "ok"

    def test_json_has_uuid_format(self):
        response = client.get("/json")
        data = response.json()
        assert isinstance(data["id"], str)
        assert len(data["id"]) == 36

    def test_json_nested_data_structure(self):
        response = client.get("/json")
        data = response.json()
        assert "items" in data["data"]
        assert "count" in data["data"]
        assert "nested" in data["data"]
        assert data["data"]["items"] == ["item1", "item2", "item3"]
        assert data["data"]["count"] == 3


class TestEchoEndpoint:
    def test_echo_returns_same_body(self):
        response = client.post("/echo", json={"key": "value"})
        assert response.status_code == 200
        assert response.json() == {"key": "value"}

    def test_echo_nested_object(self):
        payload = {"name": "test", "count": 42, "nested": {"key": "value"}}
        response = client.post("/echo", json=payload)
        assert response.status_code == 200
        assert response.json()["name"] == "test"
        assert response.json()["count"] == 42

    def test_echo_with_array(self):
        payload = {"items": [1, 2, 3]}
        response = client.post("/echo", json=payload)
        assert response.status_code == 200
        assert response.json()["items"] == [1, 2, 3]

    def test_echo_empty_body(self):
        response = client.post("/echo", json={})
        assert response.status_code == 200
        assert response.json() == {}


class TestSlowEndpoint:
    def test_slow_returns_200(self):
        response = client.get("/slow")
        assert response.status_code == 200

    def test_slow_returns_ok_status(self):
        response = client.get("/slow")
        assert response.json()["status"] == "ok"


class TestValidation:
    def test_unknown_endpoint_returns_404(self):
        response = client.get("/unknown")
        assert response.status_code == 404

    def test_post_to_get_endpoint_returns_405(self):
        response = client.post("/ping")
        assert response.status_code == 405


class TestAsyncEndpoints:
    def test_json_timestamp_is_recent(self):
        import time
        response = client.get("/json")
        data = response.json()
        timestamp = data["timestamp"] / 1000
        current_time = time.time()
        assert abs(timestamp - current_time) < 60

    def test_each_request_generates_unique_id(self):
        response1 = client.get("/json")
        response2 = client.get("/json")
        id1 = response1.json()["id"]
        id2 = response2.json()["id"]
        assert id1 != id2


def test_content_type_header():
    response = client.get("/json")
    assert "application/json" in response.headers["content-type"]
