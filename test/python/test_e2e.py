"""
End-to-end tests for lrok Python SDK.
Tests all tunnel types, API endpoints, and error handling.
"""

import os
import pytest
import requests
import time
from typing import Optional

# SDK imports will be available after generation
# from lrok import ApiClient, Configuration, TunnelsApi
# For now, we'll use requests directly

LROK_API_URL = os.environ.get("LROK_API_URL", "http://localhost:4243")
TEST_APP_URL = os.environ.get("TEST_APP_URL", "http://localhost:8000")
API_KEY = os.environ.get("LUM_API_KEY", "lum_test_key")


class LrokTestClient:
    """Test client for lrok API"""

    def __init__(self, base_url: str, api_key: Optional[str] = None):
        self.base_url = base_url
        self.api_key = api_key
        self.session = requests.Session()
        if api_key:
            self.session.headers.update({"Authorization": f"Bearer {api_key}"})

    def health(self):
        """Check health"""
        response = self.session.get(f"{self.base_url}/api/v1/health")
        response.raise_for_status()
        return response.json()

    def create_tunnel(self, **kwargs):
        """Create a tunnel"""
        response = self.session.post(
            f"{self.base_url}/api/v1/tunnels",
            json=kwargs
        )
        response.raise_for_status()
        return response.json()

    def list_tunnels(self):
        """List tunnels"""
        response = self.session.get(f"{self.base_url}/api/v1/tunnels")
        response.raise_for_status()
        return response.json()

    def get_tunnel(self, tunnel_id: str):
        """Get tunnel details"""
        response = self.session.get(f"{self.base_url}/api/v1/tunnels/{tunnel_id}")
        response.raise_for_status()
        return response.json()

    def delete_tunnel(self, tunnel_id: str):
        """Delete tunnel"""
        response = self.session.delete(f"{self.base_url}/api/v1/tunnels/{tunnel_id}")
        response.raise_for_status()
        return response.status_code == 204

    def get_stats(self, tunnel_id: str):
        """Get tunnel stats"""
        response = self.session.get(f"{self.base_url}/api/v1/tunnels/{tunnel_id}/stats")
        response.raise_for_status()
        return response.json()

    def get_requests(self, tunnel_id: str):
        """Get tunnel requests (HTTP only)"""
        response = self.session.get(f"{self.base_url}/api/v1/tunnels/{tunnel_id}/requests")
        response.raise_for_status()
        return response.json()


@pytest.fixture(scope="session")
def client():
    """Create test client"""
    return LrokTestClient(LROK_API_URL, API_KEY)


@pytest.fixture(scope="session")
def wait_for_daemon(client):
    """Wait for daemon to be healthy"""
    for _ in range(30):
        try:
            health = client.health()
            assert health['status'] == 'ok'
            print(f"✅ Daemon is healthy (version: {health.get('version', 'unknown')})")
            return
        except Exception as e:
            print(f"⏳ Waiting for daemon... {e}")
            time.sleep(1)
    pytest.fail("Daemon did not become healthy in time")


@pytest.fixture
def cleanup_tunnels(client):
    """Cleanup tunnels after test"""
    yield
    # Cleanup all tunnels after each test
    try:
        tunnels = client.list_tunnels()
        for tunnel in tunnels.get('tunnels', []):
            try:
                client.delete_tunnel(tunnel['id'])
            except:
                pass
    except:
        pass


class TestSystemEndpoints:
    """Test system endpoints"""

    def test_health_check(self, client, wait_for_daemon):
        """Test health check endpoint"""
        health = client.health()

        assert health['status'] == 'ok'
        assert 'version' in health
        assert 'uptime' in health
        assert health['uptime'] >= 0

    def test_health_no_auth_required(self):
        """Test health endpoint doesn't require authentication"""
        client_no_auth = LrokTestClient(LROK_API_URL)
        health = client_no_auth.health()
        assert health['status'] == 'ok'


class TestHTTPTunnels:
    """Test HTTP tunnel creation and management"""

    def test_create_http_tunnel_basic(self, client, wait_for_daemon, cleanup_tunnels):
        """Test creating a basic HTTP tunnel"""
        tunnel = client.create_tunnel(
            type="http",
            localPort=8000
        )

        # Verify tunnel properties
        assert tunnel['id'] is not None
        assert tunnel['type'] == 'http'
        assert tunnel['localPort'] == 8000
        assert tunnel['status'] == 'connected'
        assert tunnel['publicUrl'] is not None
        assert 'lum.tools' in tunnel['publicUrl']
        assert 'createdAt' in tunnel

        # Cleanup
        assert client.delete_tunnel(tunnel['id'])

    def test_create_http_tunnel_with_subdomain(self, client, wait_for_daemon, cleanup_tunnels):
        """Test creating HTTP tunnel with custom subdomain"""
        subdomain = f"test-py-{int(time.time())}"

        tunnel = client.create_tunnel(
            type="http",
            localPort=8000,
            subdomain=subdomain
        )

        assert tunnel['status'] == 'connected'
        assert subdomain in tunnel['publicUrl']

        # Cleanup
        client.delete_tunnel(tunnel['id'])

    def test_create_http_tunnel_with_name(self, client, wait_for_daemon, cleanup_tunnels):
        """Test creating HTTP tunnel with custom name"""
        name = f"my-tunnel-{int(time.time())}"

        tunnel = client.create_tunnel(
            type="http",
            localPort=8000,
            name=name
        )

        assert tunnel['name'] == name

        # Cleanup
        client.delete_tunnel(tunnel['id'])

    def test_http_tunnel_with_encryption(self, client, wait_for_daemon, cleanup_tunnels):
        """Test HTTP tunnel with encryption enabled"""
        tunnel = client.create_tunnel(
            type="http",
            localPort=8000,
            encryption=True
        )

        assert tunnel['status'] == 'connected'
        assert tunnel['config']['encryption'] is True

        # Cleanup
        client.delete_tunnel(tunnel['id'])

    def test_http_tunnel_with_compression(self, client, wait_for_daemon, cleanup_tunnels):
        """Test HTTP tunnel with compression enabled"""
        tunnel = client.create_tunnel(
            type="http",
            localPort=8000,
            compression=True
        )

        assert tunnel['status'] == 'connected'
        assert tunnel['config']['compression'] is True

        # Cleanup
        client.delete_tunnel(tunnel['id'])


class TestTCPTunnels:
    """Test TCP tunnel creation"""

    def test_create_tcp_tunnel(self, client, wait_for_daemon, cleanup_tunnels):
        """Test creating a TCP tunnel"""
        tunnel = client.create_tunnel(
            type="tcp",
            localPort=5432,
            remotePort=10001
        )

        assert tunnel['type'] == 'tcp'
        assert tunnel['localPort'] == 5432
        assert tunnel['remotePort'] == 10001
        assert tunnel['status'] == 'connected'

        # Cleanup
        client.delete_tunnel(tunnel['id'])

    def test_tcp_tunnel_requires_remote_port(self, client, wait_for_daemon):
        """Test that TCP tunnel requires remote port"""
        with pytest.raises(requests.exceptions.HTTPError) as exc_info:
            client.create_tunnel(
                type="tcp",
                localPort=5432
            )

        assert exc_info.value.response.status_code == 400
        assert 'remote port' in exc_info.value.response.text.lower()

    def test_tcp_tunnel_with_encryption(self, client, wait_for_daemon, cleanup_tunnels):
        """Test TCP tunnel with encryption"""
        tunnel = client.create_tunnel(
            type="tcp",
            localPort=5432,
            remotePort=10002,
            encryption=True,
            compression=True
        )

        assert tunnel['status'] == 'connected'
        assert tunnel['config']['encryption'] is True
        assert tunnel['config']['compression'] is True

        # Cleanup
        client.delete_tunnel(tunnel['id'])


class TestSTCPTunnels:
    """Test STCP tunnel creation"""

    def test_create_stcp_tunnel(self, client, wait_for_daemon, cleanup_tunnels):
        """Test creating an STCP tunnel"""
        tunnel = client.create_tunnel(
            type="stcp",
            localPort=5432,
            secretKey="test-secret-key-12345678",
            name="secure-db"
        )

        assert tunnel['type'] == 'stcp'
        assert tunnel['name'] == 'secure-db'
        assert tunnel['status'] == 'connected'

        # Cleanup
        client.delete_tunnel(tunnel['id'])

    def test_stcp_requires_secret_key(self, client, wait_for_daemon):
        """Test that STCP requires secret key"""
        with pytest.raises(requests.exceptions.HTTPError) as exc_info:
            client.create_tunnel(
                type="stcp",
                localPort=5432
            )

        assert exc_info.value.response.status_code == 400
        assert 'secret key' in exc_info.value.response.text.lower()

    def test_stcp_secret_key_min_length(self, client, wait_for_daemon):
        """Test that STCP secret key must be at least 8 characters"""
        with pytest.raises(requests.exceptions.HTTPError) as exc_info:
            client.create_tunnel(
                type="stcp",
                localPort=5432,
                secretKey="short"
            )

        assert exc_info.value.response.status_code == 400


class TestTunnelManagement:
    """Test tunnel listing and retrieval"""

    def test_list_tunnels_empty(self, client, wait_for_daemon):
        """Test listing tunnels when none exist"""
        tunnels = client.list_tunnels()

        assert 'tunnels' in tunnels
        assert isinstance(tunnels['tunnels'], list)

    def test_list_tunnels_with_tunnels(self, client, wait_for_daemon, cleanup_tunnels):
        """Test listing tunnels with active tunnels"""
        # Create a few tunnels
        tunnel1 = client.create_tunnel(type="http", localPort=8000)
        tunnel2 = client.create_tunnel(type="http", localPort=8001)

        tunnels = client.list_tunnels()

        assert len(tunnels['tunnels']) >= 2
        tunnel_ids = [t['id'] for t in tunnels['tunnels']]
        assert tunnel1['id'] in tunnel_ids
        assert tunnel2['id'] in tunnel_ids

        # Cleanup
        client.delete_tunnel(tunnel1['id'])
        client.delete_tunnel(tunnel2['id'])

    def test_get_tunnel_details(self, client, wait_for_daemon, cleanup_tunnels):
        """Test getting tunnel details"""
        tunnel = client.create_tunnel(type="http", localPort=8000)

        details = client.get_tunnel(tunnel['id'])

        assert details['id'] == tunnel['id']
        assert details['type'] == 'http'
        assert details['localPort'] == 8000

        # Cleanup
        client.delete_tunnel(tunnel['id'])

    def test_get_nonexistent_tunnel(self, client, wait_for_daemon):
        """Test getting a nonexistent tunnel returns 404"""
        with pytest.raises(requests.exceptions.HTTPError) as exc_info:
            client.get_tunnel("nonexistent-id")

        assert exc_info.value.response.status_code == 404

    def test_delete_tunnel(self, client, wait_for_daemon):
        """Test deleting a tunnel"""
        tunnel = client.create_tunnel(type="http", localPort=8000)

        # Delete tunnel
        assert client.delete_tunnel(tunnel['id'])

        # Verify it's gone
        with pytest.raises(requests.exceptions.HTTPError) as exc_info:
            client.get_tunnel(tunnel['id'])
        assert exc_info.value.response.status_code == 404

    def test_delete_nonexistent_tunnel(self, client, wait_for_daemon):
        """Test deleting nonexistent tunnel returns 404"""
        with pytest.raises(requests.exceptions.HTTPError) as exc_info:
            client.delete_tunnel("nonexistent-id")

        assert exc_info.value.response.status_code == 404


class TestTunnelStats:
    """Test tunnel statistics"""

    def test_get_tunnel_stats(self, client, wait_for_daemon, cleanup_tunnels):
        """Test getting tunnel statistics"""
        tunnel = client.create_tunnel(type="http", localPort=8000)

        stats = client.get_stats(tunnel['id'])

        assert 'bytesIn' in stats
        assert 'bytesOut' in stats
        assert 'connections' in stats
        assert 'uptime' in stats
        assert stats['uptime'] >= 0

        # Cleanup
        client.delete_tunnel(tunnel['id'])

    def test_stats_for_nonexistent_tunnel(self, client, wait_for_daemon):
        """Test getting stats for nonexistent tunnel"""
        with pytest.raises(requests.exceptions.HTTPError) as exc_info:
            client.get_stats("nonexistent-id")

        assert exc_info.value.response.status_code == 404


class TestRequestInspection:
    """Test HTTP request inspection"""

    def test_get_requests_http_tunnel(self, client, wait_for_daemon, cleanup_tunnels):
        """Test getting captured requests for HTTP tunnel"""
        tunnel = client.create_tunnel(type="http", localPort=8000)

        requests_data = client.get_requests(tunnel['id'])

        assert 'requests' in requests_data
        assert isinstance(requests_data['requests'], list)

        # Cleanup
        client.delete_tunnel(tunnel['id'])

    def test_requests_not_supported_for_tcp(self, client, wait_for_daemon, cleanup_tunnels):
        """Test that request inspection is not supported for TCP tunnels"""
        tunnel = client.create_tunnel(
            type="tcp",
            localPort=5432,
            remotePort=10001
        )

        with pytest.raises(requests.exceptions.HTTPError) as exc_info:
            client.get_requests(tunnel['id'])

        assert exc_info.value.response.status_code == 400

        # Cleanup
        client.delete_tunnel(tunnel['id'])


class TestErrorHandling:
    """Test error handling and validation"""

    def test_invalid_port_number(self, client, wait_for_daemon):
        """Test that invalid port numbers are rejected"""
        with pytest.raises(requests.exceptions.HTTPError) as exc_info:
            client.create_tunnel(type="http", localPort=999999)

        assert exc_info.value.response.status_code == 400
        assert 'port' in exc_info.value.response.text.lower()

    def test_invalid_tunnel_type(self, client, wait_for_daemon):
        """Test that invalid tunnel types are rejected"""
        with pytest.raises(requests.exceptions.HTTPError) as exc_info:
            client.create_tunnel(type="invalid", localPort=8000)

        assert exc_info.value.response.status_code == 400

    def test_missing_required_fields(self, client, wait_for_daemon):
        """Test that missing required fields are rejected"""
        with pytest.raises(Exception):
            # Should fail due to missing fields
            requests.post(
                f"{LROK_API_URL}/api/v1/tunnels",
                json={"type": "http"}  # Missing localPort
            ).raise_for_status()


class TestConcurrency:
    """Test concurrent tunnel operations"""

    def test_multiple_concurrent_tunnels(self, client, wait_for_daemon, cleanup_tunnels):
        """Test creating multiple tunnels simultaneously"""
        tunnels = []

        # Create 5 tunnels
        for i in range(5):
            tunnel = client.create_tunnel(
                type="http",
                localPort=8000 + i
            )
            tunnels.append(tunnel)

        # Verify all tunnels are in the list
        tunnel_list = client.list_tunnels()
        assert len(tunnel_list['tunnels']) >= 5

        # Cleanup
        for tunnel in tunnels:
            client.delete_tunnel(tunnel['id'])

    def test_create_delete_create(self, client, wait_for_daemon, cleanup_tunnels):
        """Test creating, deleting, and recreating tunnels"""
        for i in range(3):
            tunnel = client.create_tunnel(type="http", localPort=8000)
            assert tunnel['status'] == 'connected'

            client.delete_tunnel(tunnel['id'])

            # Verify it's deleted
            with pytest.raises(requests.exceptions.HTTPError):
                client.get_tunnel(tunnel['id'])


if __name__ == '__main__':
    pytest.main([__file__, '-v', '--tb=short'])
