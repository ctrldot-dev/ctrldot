"""Tests for tool wrapper."""

import pytest
from unittest.mock import Mock, patch
from ctrldot_crewai.tool import CtrlDotToolWrapper
from ctrldot_crewai.client import CtrlDotDeniedError


def test_tool_wrapper_deny():
    """Test that DENY raises exception."""
    mock_tool = Mock()
    mock_tool.name = "test_tool"
    mock_tool._run = Mock(return_value="result")
    
    mock_client = Mock()
    mock_client.propose_action.side_effect = CtrlDotDeniedError("DENY", "Not allowed")
    
    wrapper = CtrlDotToolWrapper(
        inner=mock_tool,
        ctrldot_client=mock_client,
        agent_id="test-agent",
        session_id="test-session"
    )
    
    with pytest.raises(CtrlDotDeniedError):
        wrapper._run(arg1="value1")
    
    # Inner tool should not be called
    mock_tool._run.assert_not_called()


def test_tool_wrapper_allow():
    """Test that ALLOW executes inner tool."""
    mock_tool = Mock()
    mock_tool.name = "test_tool"
    mock_tool._run = Mock(return_value="result")
    
    mock_client = Mock()
    mock_client.propose_action.return_value = {"decision": "ALLOW", "reason": ""}
    
    wrapper = CtrlDotToolWrapper(
        inner=mock_tool,
        ctrldot_client=mock_client,
        agent_id="test-agent",
        session_id="test-session"
    )
    
    result = wrapper._run(arg1="value1")
    
    assert result == "result"
    mock_tool._run.assert_called_once_with(arg1="value1")
