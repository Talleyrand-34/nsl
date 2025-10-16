package infrastructure

import (
	"context"
	"errors"
	"fmt"
	"strconv"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"

	ds "nsl-graph/internal/datastructs"
	e "nsl-graph/internal/repository/entities"
)

type MongoNetRepository struct {
	client     *mongo.Client
	database   *mongo.Database
	brands     *mongo.Collection
	devClasses *mongo.Collection
	zoneTypes  *mongo.Collection
	proprietaries *mongo.Collection
	zones      *mongo.Collection
	models     *mongo.Collection
	devices    *mongo.Collection
	modelPorts *mongo.Collection
	devicePorts *mongo.Collection
	connTypes  *mongo.Collection
	connections *mongo.Collection
}

func NewMongoNetRepository(uri, dbName string) (*MongoNetRepository, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	client, err := mongo.Connect(ctx, options.Client().ApplyURI(uri))
	if err != nil {
		return nil, fmt.Errorf("failed to connect to MongoDB: %w", err)
	}

	if err = client.Ping(ctx, nil); err != nil {
		return nil, fmt.Errorf("failed to ping MongoDB: %w", err)
	}

	db := client.Database(dbName)

	return &MongoNetRepository{
		client:        client,
		database:      db,
		brands:        db.Collection("brands"),
		devClasses:    db.Collection("device_classes"),
		zoneTypes:     db.Collection("zone_types"),
		proprietaries: db.Collection("proprietaries"),
		zones:         db.Collection("zones"),
		models:        db.Collection("models"),
		devices:       db.Collection("devices"),
		modelPorts:    db.Collection("model_ports"),
		devicePorts:   db.Collection("device_ports"),
		connTypes:     db.Collection("connection_types"),
		connections:   db.Collection("connections"),
	}, nil
}

func (r *MongoNetRepository) Close() error {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	return r.client.Disconnect(ctx)
}

// Brand operations
func (r *MongoNetRepository) AddBrand(brandName string) error {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	brand := e.Brand{Name: brandName}
	_, err := r.brands.InsertOne(ctx, brand)
	return err
}

func (r *MongoNetRepository) GetBrands() ([]e.Brand, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	cursor, err := r.brands.Find(ctx, bson.M{})
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var brands []e.Brand
	if err = cursor.All(ctx, &brands); err != nil {
		return nil, err
	}
	return brands, nil
}

func (r *MongoNetRepository) UpdateBrand(brandId string, newBrandName string) error {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	objID, err := primitive.ObjectIDFromHex(brandId)
	if err != nil {
		return fmt.Errorf("invalid brand ID: %w", err)
	}

	update := bson.M{"$set": bson.M{"name": newBrandName}}
	result, err := r.brands.UpdateOne(ctx, bson.M{"_id": objID}, update)
	if err != nil {
		return err
	}
	if result.MatchedCount == 0 {
		return errors.New("brand not found")
	}
	return nil
}

func (r *MongoNetRepository) DeleteBrand(brandName string) error {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	result, err := r.brands.DeleteOne(ctx, bson.M{"name": brandName})
	if err != nil {
		return err
	}
	if result.DeletedCount == 0 {
		return errors.New("brand not found")
	}
	return nil
}

// DeviceClass operations
func (r *MongoNetRepository) AddDeviceClass(deviceClassName string) error {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	devClass := e.DevClass{Name: deviceClassName}
	_, err := r.devClasses.InsertOne(ctx, devClass)
	return err
}

func (r *MongoNetRepository) GetDeviceClasses() ([]e.DevClass, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	cursor, err := r.devClasses.Find(ctx, bson.M{})
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var classes []e.DevClass
	if err = cursor.All(ctx, &classes); err != nil {
		return nil, err
	}
	return classes, nil
}

func (r *MongoNetRepository) UpdateDeviceClass(deviceClassId string, newDeviceClassName string) error {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	objID, err := primitive.ObjectIDFromHex(deviceClassId)
	if err != nil {
		return fmt.Errorf("invalid device class ID: %w", err)
	}

	update := bson.M{"$set": bson.M{"name": newDeviceClassName}}
	result, err := r.devClasses.UpdateOne(ctx, bson.M{"_id": objID}, update)
	if err != nil {
		return err
	}
	if result.MatchedCount == 0 {
		return errors.New("device class not found")
	}
	return nil
}

func (r *MongoNetRepository) DeleteDeviceClass(deviceClassName string) error {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	result, err := r.devClasses.DeleteOne(ctx, bson.M{"name": deviceClassName})
	if err != nil {
		return err
	}
	if result.DeletedCount == 0 {
		return errors.New("device class not found")
	}
	return nil
}

// ZoneType operations
func (r *MongoNetRepository) AddZoneType(zoneTypeName string) error {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	zoneType := e.ZoneType{Name: zoneTypeName}
	_, err := r.zoneTypes.InsertOne(ctx, zoneType)
	return err
}

func (r *MongoNetRepository) GetZonetypes() ([]e.ZoneType, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	cursor, err := r.zoneTypes.Find(ctx, bson.M{})
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var types []e.ZoneType
	if err = cursor.All(ctx, &types); err != nil {
		return nil, err
	}
	return types, nil
}

func (r *MongoNetRepository) UpdateZoneType(zoneTypeId string, newZoneTypeName string) error {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	objID, err := primitive.ObjectIDFromHex(zoneTypeId)
	if err != nil {
		return fmt.Errorf("invalid zone type ID: %w", err)
	}

	update := bson.M{"$set": bson.M{"name": newZoneTypeName}}
	result, err := r.zoneTypes.UpdateOne(ctx, bson.M{"_id": objID}, update)
	if err != nil {
		return err
	}
	if result.MatchedCount == 0 {
		return errors.New("zone type not found")
	}
	return nil
}

func (r *MongoNetRepository) DeleteZoneType(zoneTypeName string) error {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	result, err := r.zoneTypes.DeleteOne(ctx, bson.M{"name": zoneTypeName})
	if err != nil {
		return err
	}
	if result.DeletedCount == 0 {
		return errors.New("zone type not found")
	}
	return nil
}

// Proprietary operations
func (r *MongoNetRepository) AddProprietary(proprietaryName string) error {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	prop := e.Proprietary{Name: proprietaryName}
	_, err := r.proprietaries.InsertOne(ctx, prop)
	return err
}

func (r *MongoNetRepository) GetProperties() ([]e.Proprietary, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	cursor, err := r.proprietaries.Find(ctx, bson.M{})
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var props []e.Proprietary
	if err = cursor.All(ctx, &props); err != nil {
		return nil, err
	}
	return props, nil
}

func (r *MongoNetRepository) UpdateProprietary(proprietaryId string, newProprietaryName string) error {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	objID, err := primitive.ObjectIDFromHex(proprietaryId)
	if err != nil {
		return fmt.Errorf("invalid proprietary ID: %w", err)
	}

	update := bson.M{"$set": bson.M{"name": newProprietaryName}}
	result, err := r.proprietaries.UpdateOne(ctx, bson.M{"_id": objID}, update)
	if err != nil {
		return err
	}
	if result.MatchedCount == 0 {
		return errors.New("proprietary not found")
	}
	return nil
}

func (r *MongoNetRepository) DeleteProprietary(proprietaryName string) error {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	result, err := r.proprietaries.DeleteOne(ctx, bson.M{"name": proprietaryName})
	if err != nil {
		return err
	}
	if result.DeletedCount == 0 {
		return errors.New("proprietary not found")
	}
	return nil
}

// Zone operations
func (r *MongoNetRepository) AddZone(zoneName, fatherZoneName, fatherZoneId, proprietaryName, zoneTypeName string) error {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	var fatherID *primitive.ObjectID
	if fatherZoneId != "" {
		objID, err := primitive.ObjectIDFromHex(fatherZoneId)
		if err != nil {
			return fmt.Errorf("invalid father zone ID: %w", err)
		}
		fatherID = &objID
	}

	var propID primitive.ObjectID
	if proprietaryName != "" {
		var prop e.Proprietary
		err := r.proprietaries.FindOne(ctx, bson.M{"name": proprietaryName}).Decode(&prop)
		if err != nil {
			return fmt.Errorf("proprietary not found: %w", err)
		}
		propID = prop.ID
	}

	var zoneTypeID primitive.ObjectID
	if zoneTypeName != "" {
		var zt e.ZoneType
		err := r.zoneTypes.FindOne(ctx, bson.M{"name": zoneTypeName}).Decode(&zt)
		if err != nil {
			return fmt.Errorf("zone type not found: %w", err)
		}
		zoneTypeID = zt.ID
	}

	zone := e.Zone{
		Name:          zoneName,
		FatherZoneID:  fatherID,
		ProprietaryID: propID,
		ZoneTypeID:    zoneTypeID,
	}

	_, err := r.zones.InsertOne(ctx, zone)
	return err
}

func (r *MongoNetRepository) GetZones() ([]e.Zone, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	cursor, err := r.zones.Find(ctx, bson.M{})
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var zones []e.Zone
	if err = cursor.All(ctx, &zones); err != nil {
		return nil, err
	}
	return zones, nil
}

func (r *MongoNetRepository) UpdateZone(zoneId, newZoneName, newFatherZoneId, newZoneTypeId, newProprietaryId string) error {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	objID, err := primitive.ObjectIDFromHex(zoneId)
	if err != nil {
		return fmt.Errorf("invalid zone ID: %w", err)
	}

	update := bson.M{"$set": bson.M{}}
	
	if newZoneName != "" {
		update["$set"].(bson.M)["name"] = newZoneName
	}
	if newFatherZoneId != "" {
		fatherID, err := primitive.ObjectIDFromHex(newFatherZoneId)
		if err != nil {
			return fmt.Errorf("invalid father zone ID: %w", err)
		}
		update["$set"].(bson.M)["father_zone_id"] = fatherID
	}
	if newZoneTypeId != "" {
		ztID, err := primitive.ObjectIDFromHex(newZoneTypeId)
		if err != nil {
			return fmt.Errorf("invalid zone type ID: %w", err)
		}
		update["$set"].(bson.M)["zone_type_id"] = ztID
	}
	if newProprietaryId != "" {
		propID, err := primitive.ObjectIDFromHex(newProprietaryId)
		if err != nil {
			return fmt.Errorf("invalid proprietary ID: %w", err)
		}
		update["$set"].(bson.M)["proprietary_id"] = propID
	}

	result, err := r.zones.UpdateOne(ctx, bson.M{"_id": objID}, update)
	if err != nil {
		return err
	}
	if result.MatchedCount == 0 {
		return errors.New("zone not found")
	}
	return nil
}

func (r *MongoNetRepository) DeleteZone(zoneId string) error {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	objID, err := primitive.ObjectIDFromHex(zoneId)
	if err != nil {
		return fmt.Errorf("invalid zone ID: %w", err)
	}

	result, err := r.zones.DeleteOne(ctx, bson.M{"_id": objID})
	if err != nil {
		return err
	}
	if result.DeletedCount == 0 {
		return errors.New("zone not found")
	}
	return nil
}

// Model operations
func (r *MongoNetRepository) AddModel(modelName, brandName, deviceClassName string) error {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	var brand e.Brand
	err := r.brands.FindOne(ctx, bson.M{"name": brandName}).Decode(&brand)
	if err != nil {
		return fmt.Errorf("brand not found: %w", err)
	}

	var devClass e.DevClass
	err = r.devClasses.FindOne(ctx, bson.M{"name": deviceClassName}).Decode(&devClass)
	if err != nil {
		return fmt.Errorf("device class not found: %w", err)
	}

	model := e.ModelDevice{
		Name:          modelName,
		BrandID:       brand.ID,
		DeviceClassID: devClass.ID,
	}

	_, err = r.models.InsertOne(ctx, model)
	return err
}

func (r *MongoNetRepository) GetModels() ([]e.ModelDevice, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	cursor, err := r.models.Find(ctx, bson.M{})
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var models []e.ModelDevice
	if err = cursor.All(ctx, &models); err != nil {
		return nil, err
	}
	return models, nil
}

func (r *MongoNetRepository) UpdateModel(modelId, newModelName, newBrandId, newDeviceClassId string) error {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	objID, err := primitive.ObjectIDFromHex(modelId)
	if err != nil {
		return fmt.Errorf("invalid model ID: %w", err)
	}

	update := bson.M{"$set": bson.M{}}
	
	if newModelName != "" {
		update["$set"].(bson.M)["name"] = newModelName
	}
	if newBrandId != "" {
		brandID, err := primitive.ObjectIDFromHex(newBrandId)
		if err != nil {
			return fmt.Errorf("invalid brand ID: %w", err)
		}
		update["$set"].(bson.M)["brand_id"] = brandID
	}
	if newDeviceClassId != "" {
		dcID, err := primitive.ObjectIDFromHex(newDeviceClassId)
		if err != nil {
			return fmt.Errorf("invalid device class ID: %w", err)
		}
		update["$set"].(bson.M)["device_class_id"] = dcID
	}

	result, err := r.models.UpdateOne(ctx, bson.M{"_id": objID}, update)
	if err != nil {
		return err
	}
	if result.MatchedCount == 0 {
		return errors.New("model not found")
	}
	return nil
}

func (r *MongoNetRepository) DeleteModel(modelId string) error {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	objID, err := primitive.ObjectIDFromHex(modelId)
	if err != nil {
		return fmt.Errorf("invalid model ID: %w", err)
	}

	result, err := r.models.DeleteOne(ctx, bson.M{"_id": objID})
	if err != nil {
		return err
	}
	if result.DeletedCount == 0 {
		return errors.New("model not found")
	}
	return nil
}

// Device operations
func (r *MongoNetRepository) AddDevice(deviceLabel, modelName, zoneId, zoneName, proprietaryName string) error {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	var model e.ModelDevice
	err := r.models.FindOne(ctx, bson.M{"name": modelName}).Decode(&model)
	if err != nil {
		return fmt.Errorf("model not found: %w", err)
	}

	var zoneObjID primitive.ObjectID
	if zoneId != "" {
		zoneObjID, err = primitive.ObjectIDFromHex(zoneId)
		if err != nil {
			return fmt.Errorf("invalid zone ID: %w", err)
		}
	}

	var propID primitive.ObjectID
	if proprietaryName != "" {
		var prop e.Proprietary
		err := r.proprietaries.FindOne(ctx, bson.M{"name": proprietaryName}).Decode(&prop)
		if err != nil {
			return fmt.Errorf("proprietary not found: %w", err)
		}
		propID = prop.ID
	}

	device := e.Device{
		Label:         deviceLabel,
		ModelID:       model.ID,
		ZoneID:        zoneObjID,
		ProprietaryID: propID,
	}

	_, err = r.devices.InsertOne(ctx, device)
	return err
}

func (r *MongoNetRepository) GetDevices() ([]e.Device, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	cursor, err := r.devices.Find(ctx, bson.M{})
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var devices []e.Device
	if err = cursor.All(ctx, &devices); err != nil {
		return nil, err
	}
	return devices, nil
}

func (r *MongoNetRepository) UpdateDevice(deviceId, newDeviceLabel, newModelId, newZoneId, newProprietaryId string) error {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	objID, err := primitive.ObjectIDFromHex(deviceId)
	if err != nil {
		return fmt.Errorf("invalid device ID: %w", err)
	}

	update := bson.M{"$set": bson.M{}}
	
	if newDeviceLabel != "" {
		update["$set"].(bson.M)["label"] = newDeviceLabel
	}
	if newModelId != "" {
		modelID, err := primitive.ObjectIDFromHex(newModelId)
		if err != nil {
			return fmt.Errorf("invalid model ID: %w", err)
		}
		update["$set"].(bson.M)["model_id"] = modelID
	}
	if newZoneId != "" {
		zoneID, err := primitive.ObjectIDFromHex(newZoneId)
		if err != nil {
			return fmt.Errorf("invalid zone ID: %w", err)
		}
		update["$set"].(bson.M)["zone_id"] = zoneID
	}
	if newProprietaryId != "" {
		propID, err := primitive.ObjectIDFromHex(newProprietaryId)
		if err != nil {
			return fmt.Errorf("invalid proprietary ID: %w", err)
		}
		update["$set"].(bson.M)["proprietary_id"] = propID
	}

	result, err := r.devices.UpdateOne(ctx, bson.M{"_id": objID}, update)
	if err != nil {
		return err
	}
	if result.MatchedCount == 0 {
		return errors.New("device not found")
	}
	return nil
}

func (r *MongoNetRepository) DeleteDevice(deviceId string) error {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	objID, err := primitive.ObjectIDFromHex(deviceId)
	if err != nil {
		return fmt.Errorf("invalid device ID: %w", err)
	}

	result, err := r.devices.DeleteOne(ctx, bson.M{"_id": objID})
	if err != nil {
		return err
	}
	if result.DeletedCount == 0 {
		return errors.New("device not found")
	}
	return nil
}

// ModelPort operations
func (r *MongoNetRepository) AddModelPort(portName, positionX, positionY, modelName string) error {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	var model e.ModelDevice
	err := r.models.FindOne(ctx, bson.M{"name": modelName}).Decode(&model)
	if err != nil {
		return fmt.Errorf("model not found: %w", err)
	}

	posX, err := strconv.Atoi(positionX)
	if err != nil {
		return fmt.Errorf("invalid position X: %w", err)
	}

	posY, err := strconv.Atoi(positionY)
	if err != nil {
		return fmt.Errorf("invalid position Y: %w", err)
	}

	modelPort := e.ModelPort{
		Name:      portName,
		PositionX: posX,
		PositionY: posY,
		ModelID:   model.ID,
	}

	_, err = r.modelPorts.InsertOne(ctx, modelPort)
	return err
}

func (r *MongoNetRepository) GetModelPorts() ([]e.ModelPort, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	cursor, err := r.modelPorts.Find(ctx, bson.M{})
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var ports []e.ModelPort
	if err = cursor.All(ctx, &ports); err != nil {
		return nil, err
	}
	return ports, nil
}

func (r *MongoNetRepository) UpdateModelPort(modelPortId, newPortName, newPositionX, newPositionY, newModelId string) error {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	objID, err := primitive.ObjectIDFromHex(modelPortId)
	if err != nil {
		return fmt.Errorf("invalid model port ID: %w", err)
	}

	update := bson.M{"$set": bson.M{}}
	
	if newPortName != "" {
		update["$set"].(bson.M)["name"] = newPortName
	}
	if newPositionX != "" {
		posX, err := strconv.Atoi(newPositionX)
		if err != nil {
			return fmt.Errorf("invalid position X: %w", err)
		}
		update["$set"].(bson.M)["position_x"] = posX
	}
	if newPositionY != "" {
		posY, err := strconv.Atoi(newPositionY)
		if err != nil {
			return fmt.Errorf("invalid position Y: %w", err)
		}
		update["$set"].(bson.M)["position_y"] = posY
	}
	if newModelId != "" {
		modelID, err := primitive.ObjectIDFromHex(newModelId)
		if err != nil {
			return fmt.Errorf("invalid model ID: %w", err)
		}
		update["$set"].(bson.M)["model_id"] = modelID
	}

	result, err := r.modelPorts.UpdateOne(ctx, bson.M{"_id": objID}, update)
	if err != nil {
		return err
	}
	if result.MatchedCount == 0 {
		return errors.New("model port not found")
	}
	return nil
}

func (r *MongoNetRepository) DeleteModelPort(modelPortId string) error {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	objID, err := primitive.ObjectIDFromHex(modelPortId)
	if err != nil {
		return fmt.Errorf("invalid model port ID: %w", err)
	}

	result, err := r.modelPorts.DeleteOne(ctx, bson.M{"_id": objID})
	if err != nil {
		return err
	}
	if result.DeletedCount == 0 {
		return errors.New("model port not found")
	}
	return nil
}

// DevicePort operations
func (r *MongoNetRepository) AddDevicePort(deviceId, modelPortId, macAddress string) error {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	devID, err := primitive.ObjectIDFromHex(deviceId)
	if err != nil {
		return fmt.Errorf("invalid device ID: %w", err)
	}

	mpID, err := primitive.ObjectIDFromHex(modelPortId)
	if err != nil {
		return fmt.Errorf("invalid model port ID: %w", err)
	}

	devicePort := e.DevicePort{
		DeviceID:    devID,
		ModelPortID: mpID,
		MacAddress:  macAddress,
	}

	_, err = r.devicePorts.InsertOne(ctx, devicePort)
	return err
}

func (r *MongoNetRepository) GetDevicePorts() ([]e.DevicePort, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	cursor, err := r.devicePorts.Find(ctx, bson.M{})
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var ports []e.DevicePort
	if err = cursor.All(ctx, &ports); err != nil {
		return nil, err
	}
	return ports, nil
}

func (r *MongoNetRepository) DeleteDevicePort(deviceId, modelPortId string) error {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	devID, err := primitive.ObjectIDFromHex(deviceId)
	if err != nil {
		return fmt.Errorf("invalid device ID: %w", err)
	}

	mpID, err := primitive.ObjectIDFromHex(modelPortId)
	if err != nil {
		return fmt.Errorf("invalid model port ID: %w", err)
	}

	result, err := r.devicePorts.DeleteOne(ctx, bson.M{
		"device_id":     devID,
		"model_port_id": mpID,
	})
	if err != nil {
		return err
	}
	if result.DeletedCount == 0 {
		return errors.New("device port not found")
	}
	return nil
}

// ConnectionType operations
func (r *MongoNetRepository) AddConnectionType(connectionTypeName string) error {
	ctx
