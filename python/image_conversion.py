import argparse
import cv2
import numpy as np
from sklearn.cluster import KMeans
from skimage import io
from pathlib import Path
from PIL import Image
from rembg import remove


parser = argparse.ArgumentParser(
    prog='ProgramName',
    description='What the program does',
    epilog='Text at the bottom of help')

parser.add_argument("image")
parser.add_argument("--colours",
                    type=int,
                    default=3)
parser.add_argument("--scalar",
                    type=float,
                    default=1.0)
parser.add_argument("--output-filename",
                    type=str,
                    default=None)

args = parser.parse_args()

img = Image.open(args.image)
img = remove(img)
resized = img.resize((int(img.width * args.scalar),
                     int(img.height * args.scalar)))

arr = np.asarray(resized)
arr = arr[:, :, 0:3]
arr = arr.reshape((-1, 3))

kmeans = KMeans(n_clusters=args.colours, init='k-means++', n_init='auto', random_state=42).fit(
    arr[(arr != [0, 0, 0]).any(axis=1)])
centers = kmeans.cluster_centers_

reduced = centers[kmeans.predict(arr)]

centers = centers.astype('uint8')
reduced = reduced.astype('uint8')

reduced[arr == 0] = 0
reduced = reduced.reshape(
    (img.height, img.width, 3)).astype('uint8')

reduced_img = Image.fromarray(reduced, "RGB")

output_path = Path(args.image)
processed_path = output_path.parent / \
    (output_path.stem + f"_processed_{args.colours}" + output_path.suffix)

reduced_img.save(str(processed_path))

bitmap_dir = output_path.parent / f"{output_path.stem}_bitmaps"
bitmap_dir.mkdir(parents=True, exist_ok=True)

for center in centers:
    bitmap = reduced.copy()
    mask = np.all(bitmap == center, axis=2)
    bitmap[mask] = [255, 255, 255]
    bitmap[~mask] = [0, 0, 0]
    bitmap_img = Image.fromarray(bitmap, "RGB")
    bitmap_img.save(
        str(bitmap_dir / f"{output_path.stem}_channel_{'_'.join([str(c) for c in center])}.png"))
