# Deployment Guide - RoyalFlush Poker

This guide covers deploying the RoyalFlush poker app using:
- **Backend**: Render.com (free tier, Go)
- **Frontend**: Vercel (free tier, Angular)
- **Database**: SQLite (persistent in Render)

---

## Prerequisites

- GitHub account (for both Render and Vercel)
- Vercel account (free: https://vercel.com)
- Render account (free: https://render.com)

---

## Step 1: Deploy Backend on Render

### 1.1 Create Render Account
1. Go to https://render.com
2. Sign up with GitHub
3. Authorize GitHub access

### 1.2 Deploy Backend Service

1. Click **"New +"** → **"Web Service"**
2. Connect your GitHub repository (select `SE-RoyalFlush/Poker`)
3. Fill in deployment details:
   - **Name**: `poker-backend`
   - **Root Directory**: `Backend` (if prompted)
   - **Runtime**: `Go`
   - **Build Command**: `go build -o server ./cmd/server`
   - **Start Command**: `./server`
   - **Plan**: Free
4. Click **"Create Web Service"**

### 1.3 Configure Environment Variables

Wait for the service to be created, then:

1. Go to **Settings** → **Environment**
2. Add these variables:
   ```
   APP_ENV = production
   GO_ENV = production
   PORT = 8080
   CSRF_AUTH_KEY = <random 32-byte string>
   JWT_SECRET = <random secret string>
   FRONTEND_URL = (leave blank for now, update after Frontend deployment)
   ```

3. Generate secure keys in terminal:
   ```bash
   # For CSRF_AUTH_KEY (32 bytes = 32 chars):
   openssl rand -hex 16
   
   # For JWT_SECRET:
   openssl rand -hex 32
   ```

4. Click **"Save"** and wait for redeployment

### 1.4 Note Backend URL

Once deployed, Render will provide a URL like:
```
https://poker-backend-xxxx.onrender.com
```

Copy this URL — you'll need it for the frontend.

---

## Step 2: Deploy Frontend on Vercel

### 2.1 Create Vercel Account
1. Go to https://vercel.com
2. Sign up with GitHub
3. Authorize GitHub access

### 2.2 Deploy Frontend

1. Click **"Add New..."** → **"Project"**
2. Import your GitHub repository (`SE-RoyalFlush/Poker`)
3. Fill in deployment details:
   - **Project Name**: `poker-frontend`
   - **Framework Preset**: `Angular`
   - **Root Directory**: `frontend`
   - **Build Command**: `npm run build`
   - **Output Directory**: `dist/poker-frontend/browser`
4. Click **"Deploy"**

### 2.3 Configure Environment Variables

Before clicking "Deploy" or in **Settings** → **Environment Variables**:

1. Add this variable:
   ```
   NG_API_BASE_URL = https://poker-backend-xxxx.onrender.com
   ```
   (Replace with your actual Render backend URL)

2. Click **"Save and Redeploy"**

### 2.4 Note Frontend URL

Vercel will provide a URL like:
```
https://poker-frontend-xxxx.vercel.app
```

Copy this URL.

---

## Step 3: Link Frontend & Backend

### 3.1 Update Backend CORS

1. Go back to Render **poker-backend** service
2. Go to **Settings** → **Environment**
3. Update the `FRONTEND_URL` variable to your Vercel URL:
   ```
   FRONTEND_URL = https://poker-frontend-xxxx.vercel.app
   ```
4. Click **"Save"** (service will redeploy automatically)

### 3.2 Update Frontend API Base URL

1. Go to Vercel **poker-frontend** project
2. Go to **Settings** → **Environment Variables**
3. Set (or update) the API URL:
   ```
   NG_API_BASE_URL = https://poker-backend-xxxx.onrender.com
   ```
4. Click **"Save and Redeploy"**

---

## Step 4: Verify Deployment

1. Open your Vercel frontend URL in browser: `https://poker-frontend-xxxx.vercel.app`
2. Try **Register** → create an account → **Login**
3. Check browser console for any CORS or API errors
4. Test a few features (lobby, room creation, etc.)

---

## Troubleshooting

### CORS Errors
- Verify `FRONTEND_URL` is set correctly in Render backend
- Ensure no trailing slashes in URLs
- Check browser console for exact error message

### 404 on API Calls
- Verify `NG_API_BASE_URL` in Vercel frontend matches your Render backend URL
- Check Render logs for backend errors: **Logs** tab in Render dashboard

### SQLite Database Issues
- Render mounts the data directory at `/data`
- Database file persists between deployments **only if using persistent disk**
- To enable: Upgrade to Render's paid tier or use PostgreSQL

### Cannot Login
- Check that backend JWT_SECRET and CSRF_AUTH_KEY are set (non-empty)
- Review Render logs for application errors

---

## Database (SQLite on Render)

**Important**: With the free tier, SQLite data may not persist across redeployments.

### To Keep Data Persistent:
- **Option 1**: Upgrade to Render paid tier ($7/month) for persistent disk
- **Option 2**: Migrate to PostgreSQL:
  1. Add PostgreSQL service to Render (free tier available)
  2. Update backend `db/config.go` to use PostgreSQL driver
  3. Run migrations on PostgreSQL database

See `Backend/pkg/db/client.go` for database configuration.

---

## Updating Your Code

### To Redeploy After Code Changes:

1. **Make changes locally**
   ```bash
   git add .
   git commit -m "your message"
   git push origin main
   ```

2. **Both Render and Vercel automatically redeploy** when you push to `main`

3. **Monitor deployments**:
   - Render: **Deployments** tab
   - Vercel: **Deployments** tab

---

## Environment Variables Reference

### Backend (Render)
| Variable | Example | Notes |
|----------|---------|-------|
| `APP_ENV` | `production` | Set to production |
| `GO_ENV` | `production` | Set to production |
| `PORT` | `8080` | Render provides this automatically; set locally only when overriding the default |
| `CSRF_AUTH_KEY` | (32-byte random) | Required, should be random |
| `JWT_SECRET` | (random string) | Required for JWT signing |
| `FRONTEND_URL` | `https://poker-frontend-xxx.vercel.app` | Your Vercel frontend URL |

### Frontend (Vercel)
| Variable | Example | Notes |
|----------|---------|-------|
| `NG_API_BASE_URL` | `https://poker-backend-xxx.onrender.com` | Your Render backend URL |

---

## Monitoring & Logs

### Render (Backend)
- Go to your service dashboard
- Click **Logs** tab to see real-time logs
- Useful for debugging errors

### Vercel (Frontend)
- Go to your project dashboard
- Click **Deployments** → specific deployment → **Logs**
- Shows build output and runtime errors

---

## Next Steps

- Set up custom domain (both support free custom domains with GitHub)
- Configure automated tests in CI/CD (Render/Vercel GitHub integration)
- Monitor application metrics (both platforms provide dashboards)

**Happy Gaming!** 🎰♠️
