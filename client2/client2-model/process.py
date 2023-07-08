# import libraries here

import numpy as np
import tensorflow as tf
from tensorflow.keras.preprocessing.image import load_img, img_to_array
from tensorflow.keras.models import Sequential
from tensorflow.keras.layers import Dense, Flatten, Conv2D, MaxPooling2D
import os
import cv2



def train_or_load_gender_model(train_image_paths, train_image_labels):

    try:
        filename = 'model.h5'
        model = tf.keras.models.load_model(filename)
        print(model)
        return model
    except:
        model = None
 
    train_images = []
    
    for image_path in train_image_paths:
        #image = load_img(image_path, target_size=(50, 50))
       # print(image.shape)
        image = cv2.cvtColor(cv2.imread(image_path), cv2.COLOR_RGB2GRAY)
        image = cv2.resize(image, (50, 50))
        image = img_to_array(image) / 255.0     
        train_images.append(image)
    train_images = np.array(train_images)
    train_labels = np.array(train_image_labels)

    model = Sequential()
    model.add(Flatten(input_shape=(50, 50, 1)))
    model.add(Dense(128, activation='relu'))
    model.add(Dense(64, activation='relu'))
    model.add(Dense(1, activation='sigmoid'))

    model.compile(optimizer='adam', loss='binary_crossentropy', metrics=['accuracy'])
    model.fit(train_images, train_labels, epochs=50, batch_size=20)
    filename = 'model.h5'
    tf.keras.models.save_model(model, filename)
    return model


def getWeights(model):

    return model.get_weights()



def setWeights(weights):
    try:
        
        filename = 'model.h5'
        model = tf.keras.models.load_model(filename)
        model.layers[1].set_weights([np.array(weights[0]),np.array(weights[1])])
        model.layers[2].set_weights([np.array(weights[2]),np.array(weights[3])])
        model.layers[3].set_weights([np.array(weights[4]),np.array(weights[5])])
        
        filename = 'model.h5'
        tf.keras.models.save_model(model, filename)


    except Exception as e: 
        print(e)
        model = None
        raise Exception(e)
        







