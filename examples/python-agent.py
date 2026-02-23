#!/usr/bin/env python3
"""
Simple Python agent demonstrating Ctrl Dot integration
"""

import requests
import json
import time
import uuid

class CtrlDotAgent:
    def __init__(self, agent_id, server_url="http://127.0.0.1:7777"):
        self.agent_id = agent_id
        self.server_url = server_url
        self.session_id = None
    
    def register(self, display_name):
        """Register agent with Ctrl Dot"""
        response = requests.post(
            f"{self.server_url}/v1/agents/register",
            json={
                "agent_id": self.agent_id,
                "display_name": display_name
            }
        )
        response.raise_for_status()
        print(f"✅ Agent '{self.agent_id}' registered")
        return response.json()
    
    def start_session(self):
        """Start a session"""
        response = requests.post(
            f"{self.server_url}/v1/sessions/start",
            json={
                "agent_id": self.agent_id,
                "metadata": {"started_by": "python-agent"}
            }
        )
        response.raise_for_status()
        session = response.json()
        self.session_id = session["session_id"]
        print(f"✅ Session started: {self.session_id}")
        return session
    
    def propose_action(self, action_type, target, inputs=None, cost_gbp=0.1, resolution_token=None):
        """Propose an action to Ctrl Dot"""
        proposal = {
            "agent_id": self.agent_id,
            "session_id": self.session_id,
            "intent": {
                "title": f"Execute {action_type}"
            },
            "action": {
                "type": action_type,
                "target": target or {},
                "inputs": inputs or {}
            },
            "cost": {
                "currency": "GBP",
                "estimated_gbp": cost_gbp,
                "estimated_tokens": int(cost_gbp * 10000),
                "model": "test-model"
            },
            "context": {
                "tool": action_type,
                "tags": ["test"]
            }
        }
        
        if resolution_token:
            proposal["resolution_token"] = resolution_token
        
        response = requests.post(
            f"{self.server_url}/v1/actions/propose",
            json=proposal
        )
        response.raise_for_status()
        return response.json()
    
    def execute_action(self, action_type, target, inputs=None, cost_gbp=0.1, resolution_token=None):
        """Propose and execute action if allowed"""
        decision_resp = self.propose_action(action_type, target, inputs, cost_gbp, resolution_token)
        decision = decision_resp.get("decision")
        reason = decision_resp.get("reason", "")
        
        print(f"📋 Action: {action_type}")
        print(f"   Decision: {decision}")
        if reason:
            print(f"   Reason: {reason}")
        
        if decision in ["ALLOW", "WARN", "THROTTLE"]:
            print(f"✅ Executing action: {action_type}")
            # In real agent, execute the action here
            time.sleep(0.1)
            print(f"✅ Action completed")
            return True
        else:
            print(f"❌ Action denied: {decision}")
            return False


if __name__ == "__main__":
    print("🐍 Python Test Agent - Ctrl Dot Integration Demo")
    print("=" * 50 + "\n")
    
    # Create agent
    agent = CtrlDotAgent("python-bot")
    
    # Register
    agent.register("Python Test Bot")
    
    # Start session
    agent.start_session()
    
    print("\n--- Testing Actions ---\n")
    
    # Test 1: Git push (should be denied)
    print("Test 1: Git push action")
    agent.propose_action(
        "git.push",
        {"repo_path": "/tmp/repo", "branch": "main"},
        {"commit_message": "Test"}
    )
    
    print()
    
    # Test 2: Filesystem delete (should be denied)
    print("Test 2: Filesystem delete action")
    agent.propose_action(
        "filesystem.delete",
        {"path": "/tmp/test.txt"}
    )
    
    print("\n✅ Test complete!")
    print("\nTo use with real agents:")
    print("1. Call propose_action() before each tool execution")
    print("2. Check decision field")
    print("3. Execute only if ALLOW/WARN/THROTTLE")
    print("4. For DENY actions, request resolution token via CLI")
