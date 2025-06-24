# Magistrala Admin Dashboard

A modern, production-ready admin dashboard for managing Magistrala IoT platform. Built with React, TypeScript, and Tailwind CSS.

![Dashboard Preview](https://images.unsplash.com/photo-1551288049-bebda4e38f71?w=1200&h=600&fit=crop)

## ✨ Features

### 🔐 Authentication
- Secure login with JWT token management
- Route protection and automatic redirects
- Session persistence and logout functionality

### 📊 Dashboard Overview
- Real-time system metrics and statistics
- Service health monitoring with visual indicators
- Quick access to key system information
- Auto-refreshing data every 30 seconds

### 👥 User Management
- Complete CRUD operations for users
- Search and filter functionality
- User status tracking and management
- Secure password handling

### 📱 Device Management
- Device registration and key generation
- Device status monitoring (online/offline)
- Secure credential display and management
- Battery and connection status tracking

### 📡 Channel Management
- Create and manage data channels
- Link/unlink devices to channels
- Channel activity monitoring
- Message throughput tracking

### 🏥 System Health
- Real-time service health monitoring
- Automatic health checks every 30 seconds
- Visual status indicators for all services
- Error reporting and diagnostics

## 🛠️ Technology Stack

- **Frontend**: React 18 + TypeScript
- **Styling**: Tailwind CSS
- **Icons**: Lucide React
- **HTTP Client**: Axios
- **Routing**: React Router DOM
- **Build Tool**: Vite

## 🚀 Getting Started

### Prerequisites
- Node.js 16+ 
- npm or yarn
- Running Magistrala instance

### Installation

1. **Clone the repository**
   ```bash
   git clone <repository-url>
   cd magistrala-admin-dashboard
   ```

2. **Install dependencies**
   ```bash
   npm install
   ```

3. **Configure API endpoint**
   Edit `src/api/api.js` and update the `BASE_URL` to match your Magistrala instance:
   ```javascript
   const BASE_URL = 'http://your-magistrala-instance:9000';
   ```

4. **Start development server**
   ```bash
   npm run dev
   ```

5. **Open in browser**
   Navigate to `http://localhost:5173`

### Default Login Credentials
- **Email**: `admin@example.com`
- **Password**: `admin123`

## 📁 Project Structure

```
src/
├── api/
│   └── api.js              # API integration layer
├── components/
│   ├── Layout.jsx          # Main layout with navigation
│   ├── ProtectedRoute.jsx  # Route protection
│   ├── LoadingSpinner.jsx  # Loading component
│   ├── Modal.jsx           # Reusable modal
│   └── Toast.jsx           # Notification system
├── pages/
│   ├── Login.jsx           # Authentication page
│   ├── Dashboard.jsx       # Main dashboard
│   ├── Users.jsx           # User management
│   ├── Devices.jsx         # Device management
│   ├── Channels.jsx        # Channel management
│   └── Health.jsx          # System health monitoring
└── App.tsx                 # Main application component
```

## 🔧 Configuration

### API Endpoints
The application expects the following Magistrala API endpoints:

| Endpoint | Purpose |
|----------|---------|
| `POST /tokens` | User authentication |
| `GET /users` | List users |
| `POST /users` | Create user |
| `GET /things` | List devices |
| `POST /things` | Create device |
| `GET /channels` | List channels |
| `POST /channels` | Create channel |
| `POST /channels/{id}/things/{id}` | Link device to channel |
| `GET /{service}/health` | Health check |

### Environment Variables
Create a `.env` file for environment-specific configuration:

```env
VITE_API_BASE_URL=http://localhost:9000
VITE_APP_NAME=Magistrala Admin
```

## 🎨 Design System

### Color Palette
- **Primary**: Blue gradient (`#3B82F6` to `#8B5CF6`)
- **Success**: Green (`#10B981`)
- **Warning**: Amber (`#F59E0B`)
- **Error**: Red (`#EF4444`)
- **Gray Scale**: Neutral grays for text and backgrounds

### Typography
- **Font Family**: Inter (Google Fonts)
- **Weights**: 300, 400, 500, 600, 700
- **Scale**: Tailwind's default type scale

### Components
- Consistent 8px spacing system
- Rounded corners (8px, 12px, 16px)
- Subtle shadows and hover effects
- Responsive design breakpoints

## 🔄 API Integration

### Authentication Flow
1. User submits credentials
2. API returns JWT token
3. Token stored in localStorage
4. Token included in all subsequent requests
5. Automatic logout on token expiration

### Error Handling
- Global error interceptor
- User-friendly error messages
- Loading states for all operations
- Retry mechanisms where appropriate

### Data Management
- Optimistic updates for better UX
- Auto-refresh for critical data
- Local state management with React hooks
- Proper cleanup to prevent memory leaks

## 🧪 Development

### Available Scripts
- `npm run dev` - Start development server
- `npm run build` - Build for production
- `npm run preview` - Preview production build
- `npm run lint` - Run ESLint

### Code Style
- TypeScript for type safety
- ESLint configuration included
- Consistent component structure
- Proper prop types and interfaces

## 🚢 Deployment

### Build for Production
```bash
npm run build
```

### Deploy to Netlify/Vercel
1. Connect your repository
2. Set build command: `npm run build`
3. Set publish directory: `dist`
4. Add environment variables if needed

### Docker Deployment
```dockerfile
FROM node:18-alpine
WORKDIR /app
COPY package*.json ./
RUN npm install
COPY . .
RUN npm run build
EXPOSE 5173
CMD ["npm", "run", "preview"]
```

## 🤝 Contributing

1. Fork the repository
2. Create a feature branch
3. Make your changes
4. Add tests if applicable
5. Submit a pull request

## 📝 License

This project is licensed under the MIT License - see the [LICENSE](LICENSE) file for details.

## 🆘 Support

For support and questions:
- Open an issue on GitHub
- Check the Magistrala documentation
- Review the API integration guide

## 🗺️ Roadmap

- [ ] Real-time WebSocket integration
- [ ] Advanced charts and analytics
- [ ] Role-based access control
- [ ] Multi-tenant support
- [ ] Mobile app companion
- [ ] Advanced device management
- [ ] Custom dashboard widgets
- [ ] Export/import functionality

---

Built with ❤️ for the Magistrala community