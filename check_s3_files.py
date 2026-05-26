import os
import boto3
from dotenv import load_dotenv

def main():
    # Membaca konfigurasi dari file .env
    load_dotenv()

    endpoint = os.getenv("S3_URL")
    access_key = os.getenv("S3_ACCESS_KEY_ID")
    secret_key = os.getenv("S3_SECRET_ACCESS_KEY")
    bucket = os.getenv("S3_BUCKET")
    prefix_env = os.getenv("S3_KEY_PREFIX", "")

    if not all([endpoint, access_key, secret_key, bucket]):
        print("Error: Konfigurasi S3 di .env tidak lengkap.")
        return

    # Mensimulasikan logika prefix yang sama persis dengan kode Go
    # strings.TrimSuffix(os.Getenv("S3_KEY_PREFIX"), "/")
    # var prefix string
    # if s3KeyPrefix != "" { prefix = s3KeyPrefix + "/" }
    prefix = prefix_env.rstrip("/")
    if prefix:
        prefix += "/"

    # Boto3 butuh skema (http/https), kalau belum ada kita asumsikan https (sama seperti useSSL=true di Go)
    if not endpoint.startswith("http"):
        endpoint_url = f"https://{endpoint}"
    else:
        endpoint_url = endpoint

    print(f"Menghubungkan ke S3/MinIO: {endpoint_url}")
    print(f"Bucket: {bucket}")
    print(f"Prefix: '{prefix}'")
    print("-" * 50)

    try:
        # Inisialisasi client S3
        s3_client = boto3.client(
            's3',
            endpoint_url=endpoint_url,
            aws_access_key_id=access_key,
            aws_secret_access_key=secret_key,
            # Menambahkan config umum untuk S3-compatible API seperti MinIO
            config=boto3.session.Config(signature_version='s3v4'),
            region_name='us-east-1' # Default region dummy
        )

        # Mengambil list file dari bucket sesuai prefix
        response = s3_client.list_objects_v2(
            Bucket=bucket,
            Prefix=prefix
        )

        if 'Contents' in response:
            print(f"Ditemukan {len(response['Contents'])} file:\n")
            for obj in response['Contents']:
                print(f"[+] {obj['Key']}")
                print(f"    -> Last Modified: {obj['LastModified']}")
                print(f"    -> Size: {obj['Size'] / 1024 / 1024:.2f} MB\n")
        else:
            print("Tidak ada file yang ditemukan pada prefix ini (sudah kosong/bersih).")

    except Exception as e:
        print(f"Gagal mengambil list file: {e}")

if __name__ == "__main__":
    main()
