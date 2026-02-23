"""HTTP client for Ctrl Dot API."""

import os
import requests
from typing import Dict, Optional, Any


class CtrlDotError(Exception):
    """Base exception for Ctrl Dot errors."""
    pass


class CtrlDotDeniedError(CtrlDotError):
    """Raised when Ctrl Dot denies an action."""
    def __init__(self, decision: str, reason: str):
        self.decision = decision
        self.reason = reason
        super().__init__(f"Ctrl Dot denied action: {decision} - {reason}")


class CtrlDotClient:
    """HTTP client for Ctrl Dot daemon."""
    
    def __init__(
        self,
        base_url: Optional[str] = None,
        auth_token: Optional[str] = None
    ):
        """
        Initialize Ctrl Dot client.
        
        Args:
            base_url: Ctrl Dot daemon URL (defaults to CTRLDOT_URL env var or http://127.0.0.1:7777)
            auth_token: Optional bearer token (defaults to CTRLDOT_AUTH_TOKEN env var)
        """
        self.base_url = base_url or os.getenv("CTRLDOT_URL", "http://127.0.0.1:7777")
        self.auth_token = auth_token or os.getenv("CTRLDOT_AUTH_TOKEN")
        self.session = requests.Session()
        
        if self.auth_token:
            self.session.headers.update({
                "Authorization": f"Bearer {self.auth_token}"
            })
        
        self.session.headers.update({
            "Content-Type": "application/json"
        })
    
    def _request(self, method: str, path: str, **kwargs) -> requests.Response:
        """Make HTTP request with error handling."""
        url = f"{self.base_url}{path}"
        try:
            resp = self.session.request(method, url, **kwargs)
            resp.raise_for_status()
            return resp
        except requests.exceptions.ConnectionError:
            raise CtrlDotError(
                f"Cannot connect to Ctrl Dot daemon at {self.base_url}. "
                "Is the daemon running?"
            )
        except requests.exceptions.HTTPError as e:
            if resp.status_code == 403 or resp.status_code == 400:
                try:
                    error_data = resp.json()
                    decision = error_data.get("decision", "DENY")
                    reason = error_data.get("reason", str(e))
                    raise CtrlDotDeniedError(decision, reason)
                except ValueError:
                    raise CtrlDotError(f"HTTP {resp.status_code}: {resp.text}")
            # For 5xx, include response body so server error message is visible
            body = resp.text if resp.text else "(no body)"
            try:
                import json
                err_json = json.loads(resp.text)
                body = err_json.get("error", body)
            except (ValueError, AttributeError, TypeError):
                pass
            raise CtrlDotError(f"HTTP {resp.status_code}: {body}")
    
    def register_agent(
        self,
        agent_id: str,
        display_name: Optional[str] = None,
        default_mode: Optional[str] = None
    ) -> None:
        """
        Register an agent with Ctrl Dot.
        
        Args:
            agent_id: Unique agent identifier
            display_name: Optional display name
            default_mode: Optional default mode (normal, degraded, etc.)
        """
        payload = {"agent_id": agent_id}
        if display_name:
            payload["display_name"] = display_name
        if default_mode:
            payload["default_mode"] = default_mode
        
        self._request("POST", "/v1/agents/register", json=payload)
    
    def propose_action(self, payload: Dict[str, Any]) -> Dict[str, Any]:
        """
        Propose an action and get a decision.
        
        Args:
            payload: Action proposal payload
            
        Returns:
            Decision response with 'decision' and 'reason' fields
            
        Raises:
            CtrlDotDeniedError: If decision is DENY or STOP
        """
        resp = self._request("POST", "/v1/actions/propose", json=payload)
        decision_data = resp.json()
        
        decision = decision_data.get("decision", "DENY")
        reason = decision_data.get("reason", "")
        
        if decision in ("DENY", "STOP"):
            raise CtrlDotDeniedError(decision, reason)
        
        return decision_data
    
    def get_budget(self, agent_id: str) -> Dict[str, Any]:
        """
        Get budget status for an agent (optional endpoint).
        
        Args:
            agent_id: Agent identifier
            
        Returns:
            Budget status dict
        """
        resp = self._request("GET", f"/v1/agents/{agent_id}")
        agent_data = resp.json()
        # Return budget info if available in agent response
        return agent_data.get("budget", {})
