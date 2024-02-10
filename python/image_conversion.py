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

kmeans = KMeans(n_clusters=args.colours, n_init='auto').fit(
    arr[(arr != [0, 0, 0]).any(axis=1)])
centers = kmeans.cluster_centers_

reduced = centers[kmeans.predict(arr)]
reduced[arr == 0] = 0
reduced = reduced.reshape(
    (img.height, img.width, 3)).astype('uint8')

output_path = args.output_filename
if output_path is None:
    input_path = Path(args.image)
    output_path = str(input_path.parent /
                      (input_path.stem + f"_processed_{args.colours}{input_path.suffix}"))

img = Image.fromarray(reduced, "RGB")
img.save(output_path)
