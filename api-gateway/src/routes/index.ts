import { Router } from 'express';
import { authRoutes } from './auth';
import { jobRoutes } from './jobs';
import { licenseRoutes } from './license';
import { statsRoutes } from './stats';
import { filesRoutes } from './files';
import { repositoryRoutes } from './repositories';
import { usageRoutes } from './usage';

export const apiRouter = Router();

// Health check
apiRouter.get('/health', (req, res) => {
  res.json({
    status: 'healthy',
    timestamp: new Date().toISOString(),
    version: process.env.npm_package_version || '1.0.0'
  });
});

// Mount route modules
apiRouter.use('/auth', authRoutes);
apiRouter.use('/jobs', jobRoutes);
apiRouter.use('/license', licenseRoutes);
apiRouter.use('/stats', statsRoutes);
apiRouter.use('/files', filesRoutes);
apiRouter.use('/repositories', repositoryRoutes);
apiRouter.use('/usage', usageRoutes);

// 404 handler
apiRouter.use((req, res) => {
  res.status(404).json({
    error: 'Endpoint not found',
    path: req.path,
    method: req.method
  });
});