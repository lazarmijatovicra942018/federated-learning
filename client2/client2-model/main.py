from process import  train_or_load_gender_model , getWeights , setWeights
import sys
import os
from flask import Flask , jsonify ,json ,request
import jsonpickle


import numpy as np



app = Flask(__name__)

@app.route('/set_weights', methods=['POST'])
def set_weights():
    try:
        dto = request.json
        setWeights(dto)

        return jsonify({'message': 'Weights set successfully'})
    except Exception as e:
        return jsonify({'message': f'Error setting weights: {str(e)}'})



@app.route('/get_weights')
def hello():
   

    if len(sys.argv) > 1:
        TRAIN_DATASET_PATH = sys.argv[1]
    else:
        TRAIN_DATASET_PATH = '.' + os.path.sep + 'dataset' + os.path.sep + 'train' + os.path.sep


    labeled_gender = dict()


    with open(TRAIN_DATASET_PATH+'train_gender.csv', 'r') as file:
        lines = file.readlines()
        for index, line in enumerate(lines):
            if index > 0:
                cols = line.replace('\n', '').split(',')

                image =  cols[0].replace('\r', '')
                if (int(image) <=300):
                    continue
                if (image == "601"):
                    break
                if (len(image) == 3):
                    image = "000" + image + ".png"

                gender = cols[1].replace('\r', '')
                labeled_gender[str(image)] = float(gender)


    train_image_paths = []
    train_gender_labels = []

    for image_name in os.listdir(TRAIN_DATASET_PATH):

        if '.png' in image_name:
            train_image_paths.append(os.path.join(TRAIN_DATASET_PATH, image_name))
            train_gender_labels.append(labeled_gender[image_name])
            

    model = train_or_load_gender_model(train_image_paths, train_gender_labels)
    


    weights = getWeights(model)

    #setWeights(weights)
    
    dto = {
        'layer1_weights_matrix': weights[0].tolist(),
        'bias1': weights[1].tolist(),
        'layer2_weights_matrix': weights[2].tolist(),
        'bias2': weights[3].tolist(),
        'layer3_weights_matrix': weights[4].tolist(),
        'bias3': weights[5].tolist()
    }
    # Return the DTO as a JSON response
    return jsonify(dto)
    
   # return jsonpickle.encode([weights[0].tolist(),weights[1].tolist(),weights[2].tolist(),weights[3].tolist(),weights[4].tolist(),weights[5].tolist()])




# Run the application if the script is executed directly
if __name__ == '__main__':
    app.run(port=5001)
    