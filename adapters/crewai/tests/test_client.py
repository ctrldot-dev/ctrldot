"""Tests for Ctrl Dot client."""

import pytest
from unittest.mock import Mock, patch
from ctrldot_crewai.client import CtrlDotClient, CtrlDotDeniedError, CtrlDotError


def test_register_agent_success():
    """Test successful agent registration."""
    with patch("ctrldot_crewai.client.requests.Session") as mock_session:
        mock_resp = Mock()
        mock_resp.status_code = 200
        mock_resp.json.return_value = {}
        mock_session.return_value.post.return_value = mock_resp
        
        client = CtrlDotClient()
        client.session = mock_session.return_value
        
        # Should not raise
        client.register_agent("test-agent", "Test Agent")


def test_propose_action_deny():
    """Test that DENY decision raises CtrlDotDeniedError."""
    with patch("ctrldot_crewai.client.requests.Session") as mock_session:
        mock_resp = Mock()
        mock_resp.status_code = 200
        mock_resp.json.return_value = {
            "decision": "DENY",
            "reason": "Requires resolution"
        }
        mock_session.return_value.post.return_value = mock_resp
        
        client = CtrlDotClient()
        client.session = mock_session.return_value
        
        with pytest.raises(CtrlDotDeniedError) as exc_info:
            client.propose_action({"agent_id": "test"})
        
        assert exc_info.value.decision == "DENY"
        assert "resolution" in exc_info.value.reason


def test_propose_action_allow():
    """Test that ALLOW decision returns decision data."""
    with patch("ctrldot_crewai.client.requests.Session") as mock_session:
        mock_resp = Mock()
        mock_resp.status_code = 200
        mock_resp.json.return_value = {
            "decision": "ALLOW",
            "reason": ""
        }
        mock_session.return_value.post.return_value = mock_resp
        
        client = CtrlDotClient()
        client.session = mock_session.return_value
        
        result = client.propose_action({"agent_id": "test"})
        assert result["decision"] == "ALLOW"


def test_connection_error():
    """Test handling of connection errors."""
    with patch("ctrldot_crewai.client.requests.Session") as mock_session:
        import requests
        mock_session.return_value.post.side_effect = requests.exceptions.ConnectionError()
        
        client = CtrlDotClient()
        client.session = mock_session.return_value
        
        with pytest.raises(CtrlDotError) as exc_info:
            client.register_agent("test-agent")
        
        assert "Cannot connect" in str(exc_info.value)
