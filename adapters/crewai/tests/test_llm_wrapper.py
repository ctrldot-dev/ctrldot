"""Tests for LLM wrapper."""

import pytest
from unittest.mock import Mock, patch
from ctrldot_crewai.llm import CtrlDotLLM
from ctrldot_crewai.client import CtrlDotDeniedError


def test_llm_wrapper_deny():
    """Test that DENY raises exception."""
    mock_llm = Mock()
    mock_llm.call = Mock(return_value="response")
    
    mock_client = Mock()
    mock_client.propose_action.side_effect = CtrlDotDeniedError("DENY", "Not allowed")
    
    wrapper = CtrlDotLLM(
        model="test-model",
        provider_llm=mock_llm,
        ctrldot_client=mock_client,
        agent_id="test-agent",
        session_id="test-session"
    )
    
    with pytest.raises(CtrlDotDeniedError):
        wrapper.call("test message")
    
    # Provider LLM should not be called
    mock_llm.call.assert_not_called()


def test_llm_wrapper_allow():
    """Test that ALLOW calls provider LLM."""
    mock_llm = Mock()
    mock_llm.call = Mock(return_value="response")
    
    mock_client = Mock()
    mock_client.propose_action.return_value = {"decision": "ALLOW", "reason": ""}
    
    wrapper = CtrlDotLLM(
        model="test-model",
        provider_llm=mock_llm,
        ctrldot_client=mock_client,
        agent_id="test-agent",
        session_id="test-session"
    )
    
    result = wrapper.call("test message")
    
    assert result == "response"
    mock_llm.call.assert_called_once()


def test_llm_wrapper_throttle():
    """Test that THROTTLE reduces max_tokens."""
    mock_llm = Mock()
    mock_llm.call = Mock(return_value="response")
    
    mock_client = Mock()
    mock_client.propose_action.return_value = {
        "decision": "THROTTLE",
        "reason": "Budget threshold",
        "model_policy": "cheap"
    }
    
    wrapper = CtrlDotLLM(
        model="test-model",
        provider_llm=mock_llm,
        ctrldot_client=mock_client,
        agent_id="test-agent",
        session_id="test-session",
        cheap_model="cheap-model"
    )
    
    wrapper.call("test message", max_tokens=4096, temperature=1.0)
    
    # Check that throttled kwargs were passed
    call_kwargs = mock_llm.call.call_args[1]
    assert call_kwargs["max_tokens"] < 4096  # Should be reduced
    assert call_kwargs["temperature"] < 1.0  # Should be reduced
