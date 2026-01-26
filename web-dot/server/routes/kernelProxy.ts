import express from 'express';
import kernelProxyRoutes from './api.js';

/**
 * Legacy kernel proxy routes.
 *
 * We keep this for debugging/back-compat while v0.2 migrates the browser to GUI endpoints.
 * Browser should use /api/* (GUI), not /kernel/*.
 */
export const kernelProxy = express.Router();
kernelProxy.use('/', kernelProxyRoutes);

