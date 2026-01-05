# Deployment Guide: Google Cloud Run & GitHub Actions

Follow these steps to configure your Google Cloud project and GitHub repository for automated deployment.

## 1. Google Cloud Setup

### 1.1 Create a Project

If you haven't already, create a new project in the Google Cloud Console.

### 1.2 Enable APIs

Enable the following APIs:

- **Cloud Run API**
- **Artifact Registry API**
- **IAM Credentials API**

You can do this via the console or run this command (if you have `gcloud` installed locally):

```bash
gcloud services enable run.googleapis.com artifactregistry.googleapis.com iamcredentials.googleapis.com
```

### 1.3 Create Artifact Registry Repository

Create a Docker repository in Artifact Registry to store your container images.

1. Go to **Artifact Registry**.
2. Click **Create Repository**.
3. Name: `maukemana-backend`
4. Format: **Docker**
5. Mode: **Standard**
6. Location type: **Region**
7. Region: `asia-southeast2` (Jakarta) or your preferred region (must match `REGION` in `.github/workflows/deploy.yml`).
8. Click **Create**.

### 1.4 Create Service Account

Create a service account for GitHub Actions to use.

1. Go to **IAM & Admin** > **Service Accounts**.
2. Click **Create Service Account**.
3. Name: `github-actions-deployer`
4. Click **Create and Continue**.

### 1.5 Grant Permissions

We need to give this service account the power to deploy code.

1.  In the **Service Accounts** list, find the `github-actions-deployer` account you just created.
2.  Click on the **Pencil Icon** (Edit principal) for that row, OR if you are still in the creation flow, click **Continue** to go to the "Grant this service account access to project" step.
    - _Alternative:_ Go to the main **IAM** page (IAM & Admin > IAM), click **Grant Access** (at the top), paste the service account email (e.g., `github-actions-deployer@your-project.iam.gserviceaccount.com`).
3.  In the "Select a role" filter, search for and select these **3 roles** (you will need to click "+ ADD ANOTHER ROLE" for each one):
    - **Cloud Run Admin** (Allows deploying the actual service)
    - **Service Account User** (Allows the deployment process to "act as" the service account)
    - **Artifact Registry Writer** (Allows uploading the Docker image)
4.  Click **Save**.

### 1.6 Generate Key

1. Click the newly created service account (email address).
2. Go to the **Keys** tab.
3. Click **Add Key** > **Create new key**.
4. Select **JSON**.
5. The file will download automatically. **Keep this safe!**

## 2. GitHub Secrets Setup

Go to your GitHub repository > **Settings** > **Secrets and variables** > **Actions**.
Add the following repository secrets:

| Secret Name            | Value                                                                        |
| :--------------------- | :--------------------------------------------------------------------------- |
| `GCP_PROJECT_ID`       | Your Google Cloud Project ID (not name)                                      |
| `GCP_SA_KEY`           | Paste the **entire content** of the JSON key file you downloaded earlier.    |
| `DATABASE_URL`         | Connection string for your production database (Neon).                       |
| `CLERK_SECRET_KEY`     | Your Clerk Secret Key.                                                       |
| `JWT_SECRET`           | Secret for signing JWTs (if used alongside Clerk).                           |
| `OPENAI_API_KEY`       | OpenAI API Key (optional).                                                   |
| `R2_ACCOUNT_ID`        | Cloudflare R2 Account ID.                                                    |
| `R2_ACCESS_KEY_ID`     | Cloudflare R2 Access Key ID.                                                 |
| `R2_SECRET_ACCESS_KEY` | Cloudflare R2 Secret Access Key.                                             |
| `R2_BUCKET_NAME`       | Cloudflare R2 Bucket Name.                                                   |
| `R2_PUBLIC_URL`        | Public URL for the R2 bucket.                                                |
| `ALLOWED_ORIGINS`      | Comma-separated list of allowed origins (e.g., `https://your-frontend.com`). |

## 3. First Deployment

1. Commit and push the `.github/workflows/deploy.yml` file to `main`.
2. Go to the **Actions** tab in your GitHub repository.
3. You should see the "Deploy to Cloud Run" workflow running.

> **Note:** The first deployment might fail if the service doesn't exist yet and needs authentication configuration to allow public access.
> You may need to run this command once manually or adjust settings in Cloud Run console to **"Allow unauthenticated invocations"** if this is a public API.

To allow public access via command line:

```bash
gcloud run services add-iam-policy-binding maukemana-backend \
  --region=asia-southeast2 \
  --member=allUsers \
  --role=roles/run.invoker
```
