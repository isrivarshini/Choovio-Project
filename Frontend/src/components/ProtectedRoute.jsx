import React from 'react';
import { Navigate } from 'react-router-dom';
import { auth } from '../api/api';

const ProtectedRoute = ({ children }) => {
  return auth.isAuthenticated() ? children : <Navigate to="/login" replace />;
};

export default ProtectedRoute;