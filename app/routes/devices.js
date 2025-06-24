var express = require('express');
var router = express.Router();              // get an instance of the express Router

var Device   = require('../models/device');

// on routes that end in /devices
// ----------------------------------------------------
router.route('/')

    // create a devices (accessed at POST http://localhost:8080/devices)
    .post(function(req, res) {
        
        var device = new Device();        // create a new instance of the Bear model
        device.name = req.body.name;     // set the device's name (comes from the request)

        // save the device and check for errors
        device.save(function(err) {
            if (err)
                res.send(err);

            res.json({ message: 'Device created!' });
        });
        
    })

    // get all the devices (accessed at GET http://localhost:8080/devices)
    .get(function(req, res) {
        Device.find(function(err, devices) {
            if (err)
                res.send(err);

            res.json(devices);
        });
    });

    
// on routes that end in /devices/:device_id
// ----------------------------------------------------
router.route('/:device_id')

    // get the device with that id (accessed at GET http://localhost:8080/devices/:device_id)
    .get(function(req, res) {
        Device.findById(req.params.device_id, function(err, device) {
            if (err)
                res.send(err);
            res.json(device);
        });
    })

    // update the device with this id (accessed at PUT http://localhost:8080/devices/:device_id)
    .put(function(req, res) {

        // use our device model to find the device we want
        Device.findById(req.params.device_id, function(err, device) {

            if (err)
                res.send(err);

            device.name = req.body.name;  // update the devices info

            // save the device
            device.save(function(err) {
                if (err)
                    res.send(err);

                res.json({ message: 'Device updated!' });
            });

        })
    })

    // delete the device with this id (accessed at DELETE http://localhost:8080/devices/:device_id)
    .delete(function(req, res) {
        Device.remove({
            _id: req.params.device_id
        }, function(err, device) {
            if (err)
                res.send(err);

            res.json({ message: 'Successfully deleted' });
        });
    });

// export router module
module.exports = router;
